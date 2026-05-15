// Side-effecty wiring for the battleplan list page: placeholder display,
// hero-title fetch, FAB / new-plan button binding, duplicate-plan handler,
// and section arrangement. All DOM mutation lives here so the orchestrator
// stays focused on data flow.
(() => {
  window.appBattleplanList = window.appBattleplanList || {};
  const ns = window.appBattleplanList;
  if (!ns.render) return; // render module must load first

  const rndmod = ns.render;

  function showPlaceholder() {
    const placeholder = document.getElementById("battleplan-placeholder");
    if (!placeholder) return;
    const labelEl = document.getElementById("battleplan-placeholder-label");
    const labelText = window.appCopy
      && window.appCopy.battleplan
      && window.appCopy.battleplan.placeholder
      && window.appCopy.battleplan.placeholder.label;
    const aria = window.appCopy
      && window.appCopy.battleplan
      && window.appCopy.battleplan.list
      && window.appCopy.battleplan.list.sectionAriaLabel;
    if (labelEl && labelText) labelEl.textContent = labelText;
    if (aria) placeholder.setAttribute("aria-label", aria);
    placeholder.hidden = false;
    const fab = document.querySelector('[data-wizard-new]');
    if (fab) fab.hidden = true;
  }

  // Self-invoking: pull hero-title from settings on load. Safe to call before
  // access check — failures are silent (the hero is purely cosmetic).
  function initHeroTitle() {
    (async () => {
      try {
        const settings = await window.appAuth.apiFetch("/api/me/settings/battleplan");
        const title = settings && settings.data && settings.data.title;
        if (!title) return;
        const hero = document.getElementById("battleplan-hero");
        const target = document.getElementById("battleplan-hero-title");
        if (!hero || !target) return;
        const label = ((window.appCopy || {}).battleplan || {}).list.heroComingSoonLabel || "";
        target.textContent = `${title} `;
        const soon = document.createElement("span");
        soon.className = "dashboard-title-comingsoon";
        soon.textContent = label;
        target.appendChild(soon);
        hero.hidden = false;
      } catch (_) {
        /* hero is optional */
      }
    })();
  }

  // Set the FAB aria-label from copy. Called once at startup.
  function initFabAriaLabel() {
    const fab = document.querySelector('[data-wizard-new]');
    if (!fab) return;
    const aria = window.appCopy
      && window.appCopy.battleplan
      && window.appCopy.battleplan.fab
      && window.appCopy.battleplan.fab.newAria;
    if (aria) fab.setAttribute("aria-label", aria);
  }

  function bindNewPlanTrigger(trigger, needsOverwrite, draft) {
    if (!trigger) return;
    trigger.onclick = (event) => {
      const go = () => { window.location.href = `${window.appBattleplan.basePath()}/new/`; };
      if (needsOverwrite && draft) {
        event.preventDefault();
        window.appBattleplan.confirmOverwriteDraft(draft, go);
        return;
      }
      if (trigger.tagName !== "A") {
        event.preventDefault();
        go();
      }
    };
  }

  function wireNewButton(items) {
    const active = window.appBattleplan.findActive(items);
    const draft = window.appBattleplan.findDraft(items);
    // Overwrite dialog only when BOTH active and draft already exist —
    // creating "new" with active present saves as draft, which would
    // otherwise collide with the existing draft.
    const needsOverwrite = !!(active && draft);
    bindNewPlanTrigger(document.querySelector('[data-wizard-new]'), needsOverwrite, draft);
    bindNewPlanTrigger(document.getElementById("battleplan-active-new"), needsOverwrite, draft);

    const dupBtn = document.getElementById("battleplan-active-duplicate");
    if (dupBtn) {
      dupBtn.onclick = async (event) => {
        event.preventDefault();
        if (!active) return;
        const payload = {
          start_date: (active.start_date || "").slice(0, 10),
          duration_days: rndmod.durationDays(active),
          visibility: active.visibility,
          data: active.data || {},
        };
        await window.appBattleplan.createDraft({
          payload,
          items,
          basePath: window.appBattleplan.basePath(),
        });
      };
    }
  }

  function arrangeSections(listNode, items) {
    if (!listNode) return;
    const copy = (window.appCopy && window.appCopy.battleplan && window.appCopy.battleplan.list) || {};
    const activeSection = document.getElementById("battleplan-active-section");
    const draftSection = document.getElementById("battleplan-draft-section");
    const archiveSection = document.getElementById("battleplan-archive-section");
    const emptyNew = document.getElementById("battleplan-empty-new");
    const activeContainer = document.getElementById("battleplan-active");
    const draftContainer = document.getElementById("battleplan-draft");
    const activeLabel = document.getElementById("battleplan-active-label");
    const draftLabel = document.getElementById("battleplan-draft-label");
    const archiveLabel = document.getElementById("battleplan-archive-label");
    const activeActions = document.getElementById("battleplan-active-actions");
    const dupBtn = document.getElementById("battleplan-active-duplicate");
    const newBtn = document.getElementById("battleplan-active-new");

    const activeEl = listNode.querySelector(".list-item-spotlight");
    if (activeContainer) {
      activeContainer.innerHTML = "";
      if (activeEl) {
        activeContainer.appendChild(activeEl);
        activeContainer.hidden = false;
      } else {
        activeContainer.hidden = true;
      }
    }

    const draftEls = Array.from(listNode.querySelectorAll(".battleplan-list-item-draft"));
    if (draftContainer) {
      draftContainer.innerHTML = "";
      draftEls.forEach((el) => draftContainer.appendChild(el));
      draftContainer.hidden = draftEls.length === 0;
    }

    const hasActive = items.some((it) => it.status === "active");
    const hasDraft = items.some((it) => it.status === "draft");
    const hasOther = items.some((it) => it.status !== "active" && it.status !== "draft");

    if (activeLabel) {
      activeLabel.textContent = copy.activeSectionLabel || "";
      activeLabel.hidden = !(hasActive && copy.activeSectionLabel);
    }
    if (draftLabel) {
      draftLabel.textContent = copy.draftSectionLabel || "";
      draftLabel.hidden = !(hasDraft && copy.draftSectionLabel);
    }
    if (archiveLabel) {
      archiveLabel.textContent = copy.archiveSectionLabel || "";
      archiveLabel.hidden = !(hasOther && copy.archiveSectionLabel);
    }
    if (dupBtn) dupBtn.textContent = copy.duplicateLabel || "";
    if (newBtn) newBtn.textContent = copy.newPlanLabel || "";
    if (activeActions) activeActions.hidden = !hasActive;
    if (activeSection) activeSection.hidden = !hasActive;
    if (draftSection) draftSection.hidden = !hasDraft;
    if (archiveSection) archiveSection.hidden = !hasOther;
    // Standalone "New Plan" CTA only when no active exists (active section
    // already exposes its own New Plan button).
    if (emptyNew) {
      emptyNew.hidden = hasActive;
      emptyNew.textContent = copy.newPlanLabel || "";
      emptyNew.setAttribute("href", `${window.appBattleplan.basePath()}/new/`);
    }
  }

  ns.actions = {
    showPlaceholder,
    initHeroTitle,
    initFabAriaLabel,
    bindNewPlanTrigger,
    wireNewButton,
    arrangeSections,
  };
})();

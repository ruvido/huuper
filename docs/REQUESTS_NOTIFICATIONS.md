# Request Pipeline

This document is the reference for the active `request` flow.

It covers:

- how `settings.requests_flow` is configured
- which runtime actions the API accepts
- which roles can view or execute each step
- how notifications are routed

## Source of truth

- `settings.requests_flow` defines the ordered workflow steps.
- Each new request stores a snapshot of that flow in its own `data`.
- Requests already in progress keep using their stored snapshot even if `settings.requests_flow` changes later.

## Step schema

Each flow step supports:

- `role`
- `action`
- `label`
- `notes`
- `filter`
- `email_to`
- `telegram_message`

Supported `role` values:

- `admin`
- `assistant`
- `guardian`

Supported step `action` values in settings:

- `assign_group`
- `assign_guardian`
- `mentoring`
- `group_approved`
- `admin_approved`

Domain constraints:

- `assign_group` must use role `admin`
- `assistant` is always the assistant of a specific group, not a global request role
- before a request has a `group`, no specific assistant can be resolved for it

Supported `filter` values:

- `local`
  Used by `assign_group` to limit selectable groups to local groups.
- `group_members`
  Used by `assign_guardian` to limit selectable guardians to members of the selected group.

Supported `email_to` values:

- `admin`
- `assistant`
- `guardian`
- `candidate`

`telegram_message` is a boolean flag. When `true`, the backend sends a Telegram message to the request group associated with the request.

## Runtime action contract

The frontend and API do not submit a generic `advance` action.

They submit the explicit action for the current step:

- `set_group`
- `set_guardian`
- `set_mentoring_done`
- `set_group_approved`
- `set_admin_approved`

Additional actions outside the normal intermediate flow:

- `reject`
- `promote`

The `workflow` payload returned by request list/detail exposes:

- `pending_action`
- `pending_flow_action`
- `pending_action_label`
- `pending_role`
- `pending_action_notes`
- `required_field`
- `can_take_pending_action`
- `can_reject`
- `flow_version`

`pending_action` uses the runtime action names above, not the raw step action names from settings.

## Role visibility

Request visibility is role-based:

- `admin` can view all requests
- `assistant` can view requests assigned to groups where he is the group assistant
- `guardian` can view requests where he is the assigned guardian

Rejected requests are only visible to admins.

## Role execution

Step execution is checked against the current step in the stored flow snapshot:

- the request must not be rejected
- the request must still be on that expected step
- the authenticated actor must match the `role` configured for that step

Effective role rules:

- `admin` can execute admin steps and also override other roles
- `assistant` can execute only steps assigned to `assistant`
- `guardian` can execute only steps assigned to the request guardian

For assignment steps:

- `set_group` respects the current `assign_group` step role and its `filter`
- `set_guardian` respects the current `assign_guardian` step role and its `filter`

This means assignment permissions come from the configured flow, not from hardcoded frontend assumptions.

## Notifications

Notifications are routed from backend code and rendered from PocketBase templates.

Template kinds:

- `requests.new_request`
- `requests.assign_group`
- `requests.assign_guardian`
- `requests.mentoring`
- `requests.group_approved`
- `requests.admin_approved`

New request submissions send an email to `settings.email.admin`.

For step notifications, the backend uses the configured step action to select the matching template kind.

## Template data

Notification templates use `data` with:

- `email.subject`
- `email.body`
- `telegram.body`

Available placeholders include:

- `request_id`
- `full_name`
- `name`
- `email`
- `action`
- `action_label`
- `group_id`
- `group_name`
- `actor_name`
- `actor_email`
- `assistant_name`
- `assistant_email`
- `guardian_name`
- `guardian_email`
- `mentoring_notes`
- `data`

// Copy for the retreats module. English only, like every other copy file —
// the Italian a visitor reads on the public page comes from the retreat
// record's own fields (title, tagline, data.intro, data.faq, ...), typed by
// the admin. Nothing here may be hardcoded inline in a component: always read
// it through window.appCopy.
window.appCopy = Object.assign({}, window.appCopy || {}, {
  retreats: {
    public: {
      brand: "Realmen",
      notFound: "Retreat not found.",
      nav: {
        event: "The event",
        place: "The place",
        info: "Practical info",
        faq: "FAQ",
      },
      hero: {
        kickerFallback: "Retreat",
      },
      stats: {
        when: "When",
        where: "Where",
        duration: "Duration",
        fee: "Fee",
      },
      duration: {
        day: "day",
        days: "days",
        night: "night",
        nights: "nights",
      },
      event: {
        eyebrow: "The event",
        heading: "The programme",
      },
      place: {
        arrivalHeading: "Getting there",
        arrivalCar: "By car",
        arrivalTrain: "By train",
        arrivalCarpooling: "Car pooling",
      },
      info: {
        eyebrow: "Practical info",
        heading: "What to bring",
        packing: "What to bring",
        included: "Included in the fee",
        notIncluded: "Not included",
      },
      signup: {
        deadlinePrefix: "Registration open until",
        capacityNote: "Spots are limited.",
        depositNote: "Registration is confirmed with the deposit payment.",
        genericNote: "Save your spot for this retreat.",
        contactCta: "Contact us",
        feeRow: "Fee",
        depositRow: "Deposit",
        depositSuffix: "non-refundable",
        balanceRow: "Balance",
        balanceSuffix: "on site",
        minAgeRow: "Minimum age",
        submitMember: "Register",
        submitMemberDeposit: "Register and pay the deposit",
        submitGuest: "Register now",
        notOpen: "Registrations are not open yet for this retreat.",
        full: "This retreat is fully booked.",
        successMember: "Registration received, check your email.",
        successGuest: "Pre-registration received. We will contact you to confirm your spot.",
        choice: {
          startCta: "Register",
          question: "Are you part of a Realmen group?",
          memberCta: "Log in",
          or: "or",
          guestCta: "Continue",
        },
        guestNote: "Spots are limited. This is a pre-registration: we will contact you to confirm your place.",
        guestFields: {
          fullName: "Full name",
          fullNamePlaceholder: "Mario Rossi",
          birthYear: "Year of birth",
          birthYearPlaceholder: "1985",
          birthYearError: "Invalid year of birth.",
          provenance: "Where you come from",
          provenancePlaceholder: "City or province",
          mobile: "Phone",
          mobilePlaceholder: "+39 333 1234567",
          mobileHint: "Required: we call you to confirm the pre-registration.",
        },
        errors: {
          alreadySubmitted: "You already registered for this retreat.",
          closed: "Registrations for this retreat are closed.",
          full: "This retreat is fully booked.",
          invalidEmail: "Invalid email address.",
          missingGuestFields: "Please fill in all required fields.",
          generic: "Registration failed, please try again.",
        },
      },
      payment: {
        success: {
          title: "Payment received",
          body: "Thank you! Your registration is confirmed. You will receive an email with all the practical details.",
        },
        cancelled: {
          title: "Payment not completed",
          body: "The payment did not go through and your place is not confirmed yet. You can try again from the retreat page.",
        },
        backHome: "Back to home",
        backToRetreat: "Back to the retreat",
      },
      accept: {
        approved: {
          title: "Registration approved",
          body: "The registrant has been emailed the link to pay the deposit. Their place is held until they pay.",
        },
        active: {
          title: "Registration approved",
          body: "This retreat has no deposit, so the registrant is already in: they have been emailed the practical details and the Telegram invite.",
        },
        already: {
          title: "Already handled",
          body: "This registration had already been approved. Nothing changed and no email was sent again.",
        },
        invalid: {
          title: "Link not valid",
          body: "This approval link is unknown or expired. Approve the registration from the admin panel instead.",
        },
        failed: {
          title: "Approval failed",
          body: "Something went wrong while approving this registration. Try again from the admin panel.",
        },
        backHome: "Back to home",
        backToRetreat: "Back to the retreat",
      },
      faq: {
        eyebrow: "FAQ",
        heading: "Frequently asked questions",
      },
      // Locale used to format the retreat's dates for the reader. Matches the
      // existing public pages (see assets/js/public/event-register.js).
      dateLocale: "it-IT",
    },
  },
});

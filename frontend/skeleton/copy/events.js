window.appCopy = Object.assign({}, window.appCopy || {}, {
  events: {
    types: {
      rally: "Rally",
      call: "Call",
      meetup: "Meetup",
    },
    list: {
      title: "Events",
      futureLabel: "Upcoming",
      pastLabel: "Past",
      emptyLabel: "No events yet",
      errorMessage: "Events unavailable.",
      sectionAriaLabel: "Events",
      statusLabels: {
        upcoming: "Upcoming",
        past: "Past",
      },
      autoTitleFallback: "Event",
    },
    detail: {
      registerLabel: "I'll attend",
      unregisterLabel: "Can't make it",
      registrationConfirmed: "You're attending",
      registrationPending: "Awaiting confirmation",
      attendanceTitle: "Attendance",
      markPresent: "Present",
      markAbsent: "Absent",
      clearAttendance: "Clear",
      attendedLabel: "You attended",
      missedLabel: "You missed this one",
      registrationsTitle: "Attendees",
      pastAttendanceTitle: "Mark attendance",
      registerError: "Could not register.",
      unregisterError: "Could not unregister.",
      attendanceError: "Could not update attendance.",
      cancelDialog: {
        title: "Cancel this event?",
        confirmLabel: "Cancel event",
        cancelLabel: "Keep event",
      },
    },
    cancel: {
      scopeLabels: {
        this: "Only this occurrence",
        future: "This and future occurrences",
        all: "All occurrences in the series",
      },
    },
  },
});

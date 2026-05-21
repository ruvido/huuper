window.appCopy = Object.assign({}, window.appCopy || {}, {
  ui: {
    admin: {
      dashboard: {
        heroTemplate: "Il Branco sono {users},<br>attivi in {groups} e {regions}",
        metrics: {
          users: "Users",
          requests: "Requests",
          guardians: "Guardians",
        },
      },
      user: {
        cancelDialog: {
          title: "Set user as cancelled?",
          description: "The user will not be visible anymore in groups.",
          confirmLabel: "Cancel user",
          fallbackError: "Cancel unavailable.",
        },
      },
    },
    requests: {
      submittedEmailNotAccepted: "Request received, but the confirmation email was not sent. Check the email address.",
      submittedEmailAlertLabel: "Confirmation email not sent",
      otpText: "Enter the code we sent to your email.",
      otpLabel: "Verification code",
      changeEmail: "Change email",
      resendCode: "Resend code",
      otpButton: "Submit request",
      submitting: "Submitting...",
      otpSending: "Sending code...",
      otpChecking: "Checking code...",
      otpVerified: "Email verified.",
      otpSendError: "Unable to send verification code. Check the email address.",
      otpInvalidError: "Invalid verification code.",
      otpErrors: {
        email_exists_user: "An account with this email already exists.",
        email_exists_request: "A request with this email is already pending.",
        email_otp_recently_sent: "Please wait a minute before requesting a new code.",
        invalid_email: "Invalid email address.",
        failed_to_generate_email_otp: "Unable to generate verification code. Please try again.",
        failed_to_check_email_otp: "Unable to send verification code. Please try again.",
      },
      submitError: "Unable to submit request. Please try again.",
      rejectDialog: {
        title: "Are you sure you want to reject candidate?",
        submitLabel: "Reject",
        emptyStatus: "Write reason.",
      },
      closeMentoringDialog: {
        title: "Mentoring completed?",
        submitLabel: "Confirm",
      },
      mentoringActionLabel: "Finalize mentoring",
    },
  },
});

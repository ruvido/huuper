# Request Notifications

This project keeps request notification routing in backend code and the message content in PocketBase templates.

## Flow settings

The `settings.requests_flow` record controls workflow steps.

Supported step fields:

- `role`
- `action`
- `label`
- `notes`
- `filter`
- `email_to`
- `telegram_message`

`email_to` supports:

- `admin`
- `assistant`
- `guardian`
- `candidate`

`telegram_message` is a boolean flag. When `true`, the backend sends a Telegram message to the request group associated with the request.

## Template kinds

Use the `templates` collection with these `kind` values:

- `requests.new_request`
- `requests.assign_group`
- `requests.assign_guardian`
- `requests.mentoring`
- `requests.group_approved`
- `requests.admin_approved`

## Template data

Notification templates use `data` with these keys:

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
- `data`

## New request email

New request submissions send an email to `settings.email.admin`.

Use the `requests.new_request` template for the content.

# Eventflow Settings

`settings.eventflow` defines event types only.

Scheduling, occurrence count, list tabs, and the set of supported fields are
application capabilities. They are intentionally not configured here.

## Shape

- `types`: list of event types.
- `types[].value`: stable machine name stored on `events.type`.
- `types[].label`: UI label.
- `types[].description`: UI/help text.
- `types[].creators`: roles allowed to create this type. Supported values:
  `admin`, `assistant`.
- `types[].required`: map of event fields required for this type. Missing keys
  are treated as `false`.
- `types[].registration.enabled`: whether users can register.
- `types[].registration.approval`: whether registrations require approval.

## Supported Required Fields

- `title`: event title.
- `url`: external link, for example a call URL.
- `location`: physical location.
- `group`: group relation. When present on an event, visibility is limited to
  members of that group.
- `end_date`: end datetime, useful for multi-day events.

Unknown fields in `required` are ignored and logged by the backend.

## Always Required Outside Settings

- `type`
- `start_date`

## Hardcoded Event Capabilities

- cadence: `once`, `weekly:{day}`, `monthly:{position}-{day}`
- count options: `1`, `3`, `6`, `12`
- tabs: `upcoming`, `archived`

## Example

```json
{
  "types": [
    {
      "value": "call",
      "label": "Call",
      "description": "Online call.",
      "creators": ["admin", "assistant"],
      "required": {
        "title": false,
        "url": true,
        "location": false,
        "group": false,
        "end_date": false
      },
      "registration": {
        "enabled": true,
        "approval": false
      }
    },
    {
      "value": "meetup",
      "label": "Meetup",
      "description": "In-person meetup.",
      "creators": ["admin", "assistant"],
      "required": {
        "title": false,
        "url": false,
        "location": true,
        "group": false,
        "end_date": false
      },
      "registration": {
        "enabled": true,
        "approval": false
      }
    }
  ]
}
```

# 0014. Throttled sliding session expiry

Date: 2026-07-11
Status: Accepted

## Context

Sessions originally expired at a fixed time 30 days after login, even when the user remained
active. Moving the expiry after every authenticated request would provide exact sliding expiry,
but it would also turn every API read into a SQLite write and create avoidable contention when a
screen makes several parallel requests.

## Decision

Treat `SESSION_TTL` as the target inactivity window and renew active sessions on a throttled
schedule. For the default 720-hour TTL, a session becomes eligible for renewal 24 hours after its
previous login or renewal. The next authenticated request atomically moves `expires_at` to
`now + SESSION_TTL` and reissues the same opaque session ID with a fresh cookie `Max-Age`.

When `SESSION_TTL` is 24 hours or less, use half the TTL as the renewal interval. A conditional
update requires the previously observed expiry and an unexpired row, so concurrent requests can
produce only one successful renewal and an expired or deleted session cannot be revived.

There is no absolute session lifetime. A session can continue to renew while it remains active;
logout and inactivity still end it.

## Alternatives considered

- Keep fixed expiry — avoids renewal writes, but signs out users who are actively using the app.
- Renew every authenticated request — gives an exact inactivity window, but needlessly serializes
  SQLite writes during normal reads and parallel page loads.
- Add an absolute maximum lifetime — limits how long a stolen active cookie can persist, but also
  forces periodic reauthentication; this can be added later if the security policy requires it.

## Consequences

- (+) Regularly active users remain signed in without explicit refresh requests from the client.
- (+) The default policy performs at most one successful renewal write per active session per day.
- (+) Existing `created_at` and `expires_at` columns are sufficient; no migration is required.
- (−) The effective default inactivity window is 29–30 days rather than exactly 30 days.
- (−) Without an absolute lifetime, possession and continued use of a session cookie can extend it
  indefinitely.
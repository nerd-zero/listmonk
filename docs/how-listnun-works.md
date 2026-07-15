# How listnun works — a plain-language overview

*Written for anyone who wants to understand the product without reading code: managers, support staff, new hires outside engineering, or engineers who just want the 10,000-foot view before diving into `docs/plan.md`.*

## What listnun actually is

listnun lets a company send email newsletters and campaigns — the same job tools like Mailchimp do — but gives every customer their own **private workspace**, not a shared account. Under the hood, each workspace runs on **Listmonk**, a well-regarded open-source newsletter tool. Listmonk normally has to be installed separately for every customer (its own server, its own database). listnun's job is to make that experience feel instant and self-service instead: a customer signs up, and thirty seconds later they have their own working workspace, with no server to set up and no software to install.

Think of it like a serviced office building. Every tenant gets their own locked office (their own private workspace, their own data, their own login) — but the building's electricity, security, and cleaning are shared infrastructure the landlord runs once for everybody. listnun is the landlord: it runs one shared, upgraded copy of Listmonk that's been taught to keep every customer's data walled off from every other customer's, so it can serve hundreds of private "offices" without provisioning hundreds of separate buildings.

## The three things a customer needs, and who provides them

Sending a marketing email that actually lands in an inbox — not a spam folder — requires three separate ingredients. listnun brings all three together automatically:

1. **A place to build and send campaigns.** This is Listmonk itself: the screen where someone writes a newsletter, manages their subscriber list, and hits send.
2. **A mail carrier that will actually deliver it.** listnun uses **Postmark**, a company that specializes in this. Every workspace gets its own private Postmark "server" — its own dedicated sending reputation, so one customer's spam complaints never affect anyone else's deliverability.
3. **Proof to inboxes that the mail is legitimate.** This is the technical handshake (DKIM, if you've heard that term) that tells Gmail/Outlook/etc. "yes, this really came from who it says it did." listnun walks the customer through setting this up for their own domain, or offers a ready-made shared domain for anyone who doesn't have one yet.

## The customer's journey, step by step

1. **Sign up.** A person creates an account and, on first login, automatically gets their own **organization** — the umbrella that everything else (their workspaces, their teammates) lives under. Think of an organization as the customer's company account.
2. **Create a workspace.** Inside their organization, they create a workspace (internally called an "instance") — this is their own private copy of Listmonk. Behind the scenes this takes a few seconds: listnun registers a new tenant in the shared Listmonk deployment and creates their dedicated Postmark server.
3. **Log in for the first time.** A one-time setup link is generated so the new workspace's admin can set their own password. (It can't just be emailed to them automatically — at this exact moment their workspace has no working "send email" configuration yet, which is the next step.)
4. **Connect a way to send mail.** The customer either brings their own domain and proves they own it (a short DNS setup, with copy-paste-ready instructions in the dashboard), verifies a single sending address instead, or opts into a shared domain listnun already owns and has pre-verified — whichever fits how much of their own infrastructure they want to touch.
5. **Send.** From here on, it behaves like any other newsletter tool — because it is one. The customer never sees or thinks about any of the shared infrastructure underneath.

Optionally, a customer can also point their **own web address** at their workspace (e.g. `mail.theircompany.com` instead of `theircompany.listnun.app`) — this is the "custom domains" feature, covered below.

## What "workspace" actually means underneath

Every workspace a customer creates has a few things tied to it, all created and torn down together:

- Its own space inside the shared Listmonk deployment — its subscribers, campaigns, and settings never mix with anyone else's, even though they're technically stored in one shared database.
- Its own dedicated Postmark sending server.
- A verified way to send mail (its own domain, a single verified address, or the shared fallback domain).
- Optionally, its own custom web address.

Deleting a workspace tears down all of the above together — it's designed so nothing gets orphaned or left half-configured, whether a customer deletes their own workspace or a platform admin removes it on their behalf.

## Custom domains, in plain terms

Normally a workspace is reachable at an address listnun assigns, like `acme.listnun.app`. Some customers want their own address instead — `mail.acme.com` — so the tool feels like it's fully theirs, not a subdomain of ours.

Making that work safely (with a valid security certificate, without exposing our servers' real internet address) leans on a service called **Cloudflare**, which specializes in exactly this problem. The customer doesn't need a Cloudflare account of their own — they just add one line to their own DNS settings (wherever their domain is registered, any provider), pointing it at an address of ours. Cloudflare handles the rest: proving the customer really owns that domain, then issuing and renewing the security certificate automatically.

**Status:** this feature is fully built on our side, but the final connection to Cloudflare's live service is still being switched on — it needs a real Cloudflare account credential that hasn't been added yet. Everything up to that point (the screens, the database records, the underlying plumbing) has already been tested against real data; only the very last "actually talk to Cloudflare" step is pending.

## Who can do what

- **Anyone with a login** can be part of one or more organizations.
- **A member** of an organization can create, view, and manage that organization's workspaces (including sending domains, custom domains, and deleting a workspace).
- **An owner** of an organization can additionally invite new teammates into it.
- **A platform admin** (listnun's own staff, not a customer) can see and manage every organization and every workspace on the platform — for support and operational purposes — from a separate admin console.

## Safety nets already in place

- **Suspend, don't just delete.** A workspace can be temporarily suspended (e.g. for a billing issue) without losing any of its data, and reactivated later.
- **Confirmations on anything irreversible.** Deleting a workspace, a sending domain, or a custom domain always asks for confirmation first and clearly states that it can't be undone.
- **Nothing is provisioned halfway on purpose.** If a step fails partway through (say, hitting an account limit with the mail carrier), the system is built to leave things in a clean, retryable state rather than a broken one — this has been deliberately tested, not just assumed.

## Where to go for more detail

- `docs/plan.md` — the original technical architecture and build plan (engineering audience).
- `docs/custom-domains.md` — the design behind the custom-domains feature described above (engineering audience).
- This document — kept in sync with what's actually built, not just what's planned. If something here stops matching reality, treat the code as the source of truth and flag the mismatch.

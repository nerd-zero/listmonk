const MAX_SLUG_LENGTH = 63;

function slugifyPart(input: string): string {
  return input
    .toLowerCase()
    .normalize("NFKD")
    .replace(/\p{Diacritic}/gu, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

// Matches the backend's reSlug pattern (^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$)
// -- combines the org name in so the same workspace name can be reused
// across orgs despite instances.slug being a single global namespace (see
// internal/provisioning's CreateInstance doc comment).
export function instanceSlug(orgName: string, instanceName: string): string {
  return [slugifyPart(orgName), slugifyPart(instanceName)]
    .filter(Boolean)
    .join("-")
    .slice(0, MAX_SLUG_LENGTH)
    .replace(/-+$/, "");
}

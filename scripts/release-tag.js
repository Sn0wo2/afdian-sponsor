const { execFileSync } = require("node:child_process")

const TAG_PATTERN = /^v(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$/

function git(...args) {
  try {
    return execFileSync("git", args, { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim()
  } catch (error) {
    const details = error.stderr?.toString().trim()
    throw new Error(details ? `\`git ${args.join(" ")}\` failed: ${details}` : `\`git ${args.join(" ")}\` failed`)
  }
}

function gitSucceeds(...args) {
  try {
    execFileSync("git", args, { stdio: "ignore" })
    return true
  } catch {
    return false
  }
}

function parseTag(tag) {
  const match = TAG_PATTERN.exec(tag)
  if (!match) throw new Error(`Invalid tag "${tag}"; expected vMAJOR.MINOR.PATCH, e.g. v1.13.0`)
  return { major: Number(match[1]), minor: Number(match[2]), patch: Number(match[3]), prerelease: match[4] ?? "" }
}

function compareTags(a, b) {
  for (const part of ["major", "minor", "patch"]) {
    if (a[part] !== b[part]) return a[part] - b[part]
  }
  if (a.prerelease === b.prerelease) return 0
  if (!a.prerelease) return 1
  if (!b.prerelease) return -1
  return a.prerelease < b.prerelease ? -1 : 1
}

function latestTag() {
  return (
    git("tag", "--list", "v*.*.*", "--sort=-v:refname")
      .split("\n")
      .filter((tag) => TAG_PATTERN.test(tag))[0] ?? ""
  )
}

function defaultBranch() {
  try {
    return git("symbolic-ref", "--short", "refs/remotes/origin/HEAD").replace(/^origin\//, "")
  } catch {
    return "main"
  }
}

async function main() {
  const tag = process.argv[2]
  if (!tag) throw new Error("Usage: make release-tag TAG=v1.13.0 (or node scripts/release-tag.js <tag> [ref])")
  const ref = process.argv[3] ?? "HEAD"
  const version = parseTag(tag)
  const major = `v${version.major}`
  const message = process.env.TAG_MESSAGE ?? "meow"

  const branch = defaultBranch()
  const previous = latestTag()
  const isNewer = !previous || compareTags(version, parseTag(previous)) > 0

  git("fetch", "--quiet", "origin", branch)
  if (!gitSucceeds("merge-base", "--is-ancestor", ref, `origin/${branch}`)) {
    throw new Error(`${ref} is not reachable from origin/${branch}; push it before tagging`)
  }

  const localExists = gitSucceeds("rev-parse", "-q", "--verify", `refs/tags/${tag}`)
  if (localExists && git("rev-list", "-n", "1", tag) !== git("rev-list", "-n", "1", ref)) {
    throw new Error(`Tag ${tag} already exists locally at a different commit; remove it with: git tag -d ${tag}`)
  }
  if (!isNewer && previous !== tag) {
    throw new Error(`${tag} is not newer than the latest tag ${previous}`)
  }

  if (localExists) {
    console.log(`Reusing existing local tag ${tag}`)
  } else {
    if (git("ls-remote", "origin", `refs/tags/${tag}`)) {
      throw new Error(`Tag ${tag} already exists on origin; fetch it first with: git fetch origin tag ${tag}`)
    }
    console.log(`Previous tag: ${previous || "(none)"}`)
    console.log(`Creating annotated tag ${tag} at ${git("rev-parse", "--short", ref)}...`)
    git("tag", "-a", tag, "-m", message, ref)
  }

  console.log(`Pushing ${tag} and updating ${major} on origin...`)
  try {
    execFileSync("git", ["push", "origin", `refs/tags/${tag}`, `+refs/tags/${tag}:refs/tags/${major}`], { stdio: "inherit" })
  } catch {
    throw new Error(`Failed to push ${tag} to origin`)
  }

  git("fetch", "--force", "--quiet", "origin", `refs/tags/${major}:refs/tags/${major}`)
  console.log(`Done: ${tag} released and ${major} now points at ${git("rev-list", "-n", "1", tag)}`)
}

main().catch((error) => {
  console.error(error.message)
  process.exitCode = 1
})

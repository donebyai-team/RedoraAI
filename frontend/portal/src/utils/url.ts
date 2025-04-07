// isActivePathSegment is similar to isActivePath, but it always check if against starts with path.
export const isActivePathSegment = (path: string, against: string | null): boolean => {
  return (against ?? '').startsWith(path)
}

export const isActivePath = (path: string, against: string | null): boolean => {
  if (path === '/' && against !== path) {
    return false
  }

  const pathSegments = path.split('/')
  if (pathSegments.length === 2) {
    return path === against
  }

  return against?.startsWith(path) ?? false
}

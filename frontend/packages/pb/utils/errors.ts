import { ConnectError, Code } from '@connectrpc/connect'

export function maybeConnectError(err: unknown): ConnectError | undefined {
  if (err == null) {
    return undefined
  }

  // Ok at some point in one project, in an earlier version of connectrpc we
  // had the problem that `err instanceof ConnectError` was returning false
  // but the constructor.name was correctly kept to ConnectError so we
  // were comparing on `constructor.name` instead of `instanceof`.
  //
  // Now today I got the exact opposite problem, the constructor was "minified"
  // to '_t' but the instanceof worked correctly.
  //
  // We are going now the following rules to extract a ConnectErrorCode:
  if (err instanceof ConnectError) {
    return err
  }

  if (err.constructor.name === 'ConnectError') {
    return err as ConnectError
  }

  return undefined
}

export function connectErrorCode(err: unknown): Code | undefined {
  return maybeConnectError(err)?.code
}

export function isConnectError(err: unknown): boolean {
  return maybeConnectError(err) != null
}

export function isNotFoundConnectRpcError(err: unknown): boolean {
  const code = connectErrorCode(err) ?? Code.Unknown

  return code === Code.NotFound
}

export function errorToMessage(err: unknown): string {
  if (err == null) {
    return ''
  }

  const connectError = maybeConnectError(err)
  if (connectError != null) {
    if (connectError.code == Code.DeadlineExceeded) {
      // Let's also print the full error message in the console, might prove useful
      console.error('DeadlineExceeded error', connectError.rawMessage)
      return 'Request timed out, the server is taking too long to respond'
    }

    return `${codeToDisplayName[connectError.code]}: ${connectError.rawMessage}`
  }

  return err.toString()
}

const codeToDisplayName: Record<Code, string> = {
  [Code.Canceled]: 'Canceled',
  [Code.Unknown]: 'Unknown',
  [Code.InvalidArgument]: 'Invalid Argument',
  [Code.DeadlineExceeded]: 'Deadline Exceeded',
  [Code.NotFound]: 'Not Found',
  [Code.AlreadyExists]: 'Already Exists',
  [Code.PermissionDenied]: 'Permission Denied',
  [Code.ResourceExhausted]: 'Resource Exhausted',
  [Code.FailedPrecondition]: 'Failed Precondition',
  [Code.Aborted]: 'Aborted',
  [Code.OutOfRange]: 'Out Of Range',
  [Code.Unimplemented]: 'Unimplemented',
  [Code.Internal]: 'Internal',
  [Code.Unavailable]: 'Unavailable',
  [Code.DataLoss]: 'DataLoss',
  [Code.Unauthenticated]: 'Unauthenticated'
}

import { DescMessage, DescService } from '@bufbuild/protobuf'
import { PromiseClient } from '@connectrpc/connect'
import { MethodInfoUnary } from '@connectrpc/connect/dist/esm/types'
import useSWR, { SWRConfiguration } from 'swr'

type UnaryMethods<S extends DescService> = {
  [M in keyof S['method']]: S['method'][M] extends MethodInfoUnary<DescMessage, DescMessage> ? M : never
}[keyof S['method']]

/**
 *
 * This is a hook that wraps `useSWR` to make it easier to use with gRPC services. Invoked like this:
 *
 * ```tsx
 * const { data, error, isLoading } = useGrpcSwr(conversationClient, 'getMessageSources', {
 *   userId: name
 * })
 * ```
 *
 * Where `conversationClient` is a gRPC client you created with `createPromiseClient` and the
 * literal `'getMessageSources'` is the name of the method you want to call on the client if you
 * were doing `conversationClient.getMessageSources({ userId: name })`.
 *
 * Method names are inferred so Ctrl-<space> should give you the list of methods available on the client.
 * And after that args are validated against the method signature correctly.
 */
export const useGrpcSwr = <
  S extends DescService,
  M extends UnaryMethods<S>,
  P extends Parameters<PromiseClient<S>[M]>,
  R extends ReturnType<PromiseClient<S>[M]>
>(
  service: PromiseClient<S>,
  method: M,
  request: P[0],
  options: SWRConfiguration<Awaited<R>, unknown, () => Promise<R>> = {}
) => {
  return useSWR<Awaited<R>, unknown, string>(
    `${service.constructor.name}/${String(method)}${requestToString(request)}`,
    // @ts-ignore I didn't find the right way so far to type this all right, for now it's fine
    async () => service[method](request),
    options
  )
}

// We don't want to use `toJsonString` from `@bufbuild/protobuf` because it requires the usage of the
// Protobuf schema associated with the request object which we find cumbersome to use in the hook above.
//
// The thing with the `request` object is that it's a Protobuf object and can contains `BigInt` values
// that are not handled by `JSON.stringify`. So we need to convert them to strings before serializing
// as well as maybe any other object that would require it.
export function requestToString(request: unknown) {
  return JSON.stringify(request, function (this: unknown, _, value) {
    return this instanceof BigInt ? value.toString() : value
  })
}

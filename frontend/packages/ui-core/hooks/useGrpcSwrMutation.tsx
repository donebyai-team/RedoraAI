import { DescMessage, DescService } from '@bufbuild/protobuf'
import { PromiseClient } from '@connectrpc/connect'
import { MethodInfoUnary } from '@connectrpc/connect/dist/esm/types'
import useSWRMutation, { SWRMutationConfiguration } from 'swr/mutation'

type UnaryMethods<S extends DescService> = {
  [M in keyof S['method']]: S['method'][M] extends MethodInfoUnary<DescMessage, DescMessage> ? M : never
}[keyof S['method']]

/**
 *
 * This is a hook that wraps `useSWRMutation` to make it easier to use with gRPC services. Invoked like this:
 *
 * ```tsx
 * const { trigger: saveQuote, loading: isSavingQuote } = useSWRMutation(conversationClient.saveQuote)
 *
 * saveQuote({ quoteId: id }).catch((err: unknown) => { })
 * ```
 *
 * Where `conversationClient` is a gRPC client you created with `createPromiseClient` and the
 * literal saveQuote is the method you want to call on the client if you were doing
 * `conversationClient.saveQuote({ userId: name })` directly.
 */
export const useGrpcSwrMutation = <
  S extends DescService,
  M extends UnaryMethods<S>,
  P extends Parameters<PromiseClient<S>[M]>,
  R extends ReturnType<PromiseClient<S>[M]>
>(
  service: PromiseClient<S>,
  method: M,
  options: SWRMutationConfiguration<Awaited<R>, unknown, () => Promise<R>> = {}
) => {
  return useSWRMutation<Awaited<R>, unknown, string, P[0]>(
    `${service.constructor.name}/${String(method)}`,
    // @ts-ignore I didn't find the right way so far to type this all right, for now it's fine
    async (_key: string, { arg }: { arg: P[0] }) => {
      // @ts-ignore
      return service[method](arg)
    },
    options
  )
}

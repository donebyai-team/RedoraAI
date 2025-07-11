'use client'

import { routes } from '@doota/ui-core/routing'
import { FallbackSpinner } from '../../../../atoms/FallbackSpinner'
import { useSearchParams } from 'next/navigation'
import { useEffect } from 'react'
import { portalClient } from '../../../../services/grpc'
import { useRouter } from 'next/navigation'
import { log } from '../../../../services/logger'
import toast from 'react-hot-toast'
import { ConnectError } from "@connectrpc/connect";


export default function Page() {
  const searchParams = useSearchParams()
  const router = useRouter()

  useEffect(() => {
    const handleCallback = async () => {
      // If its a microsoft callback, we get back a tenant and code is null
      // If its a google callback, we get back a code and tenant is null
      const { code, tenant, stateHash, error } = {
        code: searchParams.get('code') ?? undefined,
        tenant: searchParams.get('tenant'),
        stateHash: searchParams.get('state'),
        error: searchParams.get('error')
      }
      // we always should have a state but if we don't, we raise an error
      if (!stateHash) {
        throw new Error('No state hash found')
      }
      if (error || (!code && !tenant)) {
        console.error('callback error: ', error ?? 'no code')
      }

      log.info('callback', { code, tenant, stateHash, error })
      const res = await portalClient.oauthCallback(
        {
          state: stateHash,
          externalCode: tenant ?? code
        },
        {
          timeoutMs: 30000
        }
      )
      log.info('callback answer', res)

      router.push(res.redirectUrl)
    }

    handleCallback().catch((err) => {
      let userMessage = "Something went wrong";

      if (err instanceof ConnectError) {
        // Strip out the [code] prefix and GRPC-style `desc = ` part
        const match = err.message.match(/desc = (.+)$/);
        if (match) {
          userMessage = match[1]; // Clean, user-friendly message
        } else {
          // fallback to raw message without `[code]` prefix
          userMessage = err.message.replace(/^\[\w+_?\w*\]\s*/, '');
        }
      } else if (err?.response?.data?.message) {
        userMessage = err.response.data.message;
      } else if (err?.message) {
        userMessage = err.message;
      }

      toast.error(userMessage);
      console.error("callback error: ", err);
      router.push(routes.new.dashboard);
    });


    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams])

  return <FallbackSpinner />
}

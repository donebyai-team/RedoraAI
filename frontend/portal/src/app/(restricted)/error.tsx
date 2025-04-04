'use client'

import { UnexpectedError } from '@doota/ui-core/components/UnexpectedError'
import { useRouter } from 'next/navigation'

export default function Error({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  const { push } = useRouter()

  // The reset below is coming from Next.js framework directly and will attempt to re-render the segment
  return (
    <UnexpectedError title='Unexpected client-side error' error={error} backToHome={() => push('/')} retry={reset} />
  )
}

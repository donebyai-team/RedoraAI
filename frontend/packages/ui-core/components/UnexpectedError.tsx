'use client'

import { FC, useEffect, useState } from 'react'
import { ServerCrash } from 'lucide-react'
import { PrimaryButtonPush } from '../atoms/ButtonPush'
import { errorToMessage } from '@doota/pb/utils/errors'

const hasClipboardSupport = typeof navigator !== 'undefined' && navigator.clipboard != null

type Props = {
  title?: string
  error: Error & { digest?: string }
  retry?: () => void
  backToHome?: () => void
}

/**
 * Full page error component for unexpected errors client side. Used in Portal within a Next.js
 * error boundary (`app/dashboard/error.tsx`) and within the standalone panel to display a better
 * ui when something when wrong.
 */
export const UnexpectedError: FC<Props> = ({ title, error, backToHome, retry }) => {
  const [showDetails, setShowDetails] = useState(false)

  useEffect(() => {
    // FIXME There was an error in the application, should we log it somewhere?
    console.error('An unexpected error was caught', error)
  }, [error])

  const message = `${errorToMessage(error)}\n${error.stack?.split('\n').slice(1).join('\n')}`

  const copyToClipboard = () => {
    navigator.clipboard.writeText(message).catch(err => {
      console.error('Failed to copy error details to clipboard', err)
    })
  }

  return (
    <div className='flex flex-col w-full min-h-screen items-center justify-center'>
      <div className='flex flex-col items-center'>
        <h1 className='font-extrabold text-8xl mb-1'>500</h1>
        <h5>{title ?? 'Unexpected error'} 👨🏻‍💻</h5>
        <p className='text-xs'>Oops, something went wrong!</p>
        <button
          className='text-xs underline cursor-pointer'
          onClick={() => {
            setShowDetails(!showDetails)
          }}
        >
          {showDetails ? 'Hide details' : 'Show details'}
        </button>

        {showDetails && (
          <div className='flex flex-col'>
            {hasClipboardSupport && (
              <button className='text-xs underline cursor-pointer mt-2' onClick={copyToClipboard}>
                Copy to clipboard
              </button>
            )}

            <div className='bg-gray-100 p-4 rounded-lg'>
              <pre className='whitespace-pre-wrap overflow-x-auto'>
                {errorToMessage(error)}
                {'\n'}
                {error.stack?.split('\n').slice(1).join('\n')}
              </pre>
            </div>
          </div>
        )}
      </div>

      <ServerCrash size={100} className='m-3' />
      {(backToHome || retry) && (
        <div className='flex gap-2'>
          {retry && (
            <PrimaryButtonPush className='!w-[150px]' onClick={retry}>
              Retry
            </PrimaryButtonPush>
          )}
          {backToHome && (
            <PrimaryButtonPush className='!w-[150px]' onClick={backToHome}>
              Back to Home
            </PrimaryButtonPush>
          )}
        </div>
      )}
    </div>
  )
}

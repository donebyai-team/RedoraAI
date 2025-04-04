import { ReactNode } from 'react'
import { useIsExecutionRuntimeInPortal } from '../hooks/useExecutionRuntime'

type props = {
  header: ReactNode
  body: ReactNode
  footer: ReactNode
}
export function SidePanel({ header, body, footer }: props) {
  const isInPortal = useIsExecutionRuntimeInPortal()

  return (
    <div className='relative h-full'>
      <div
        className={
          'relative h-full shadow bg-white flex flex-col gap-0 z-10' +
          (isInPortal ? ' border-l border-solid border-neutral-200' : '')
        }
      >
        {header && (
          <div
            className={
              'px-3.5 flex-shrink-0 flex justify-between items-center border-b border-neutral-200 text-black min-h-[45px] empty:hidden'
            }
          >
            {header}
          </div>
        )}
        <div className='flex-1 overflow-y-scroll'>{body}</div>
        {footer && (
          <div
            id='load-card-footer'
            className='p-3.5 flex-shrink-0 border-t border-gray-2 flex items-center justify-center'
          >
            <div className='flex flex-col gap-[5px] items-center justify-center w-full'>{footer}</div>
          </div>
        )}
      </div>
    </div>
  )
}

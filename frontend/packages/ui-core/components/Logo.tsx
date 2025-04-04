import { FC } from 'react'
import { Image } from '../atoms/Image'
import { useIsExecutionRuntimeInPortal } from '../hooks/useExecutionRuntime'

export const Logo: FC = () => {
  const isInPortal = useIsExecutionRuntimeInPortal()
  if (isInPortal) {
    return (
      <div className='flex flex-col items-center gap-4'>
        <Image width={200} height={150} alt='doota logo' priority imageKey='logo_new' />
        <div className='!text-[24px] font-bold text-[#284150]'>doota</div>
      </div>
    )
  }

  return null
}

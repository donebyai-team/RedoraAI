import { ArrowLeft } from 'lucide-react'
import { FC } from 'react'

export const PaginationBackButton: FC<{ disabled: boolean }> = ({ disabled }) => {
  return (
    <button
      disabled={disabled}
      className='bg-neutral-100 mr-1 hover:bg-neutral-200 h-5 w-5 justify-center items-center rounded-full flex'
    >
      <ArrowLeft size={14} />
    </button>
  )
}

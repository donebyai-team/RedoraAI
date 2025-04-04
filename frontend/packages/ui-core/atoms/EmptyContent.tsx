import { CircleSlash2 } from 'lucide-react'
import { FC } from 'react'

type Props = {
  text?: string
  className?: string
}

export const EmptyContent: FC<Props> = ({ text, className }) => {
  return (
    <div className={`flex flex-col items-center gap-1 ${className}`}>
      <CircleSlash2 className='text-gray-3' size={32} />
      {text && <div className='text-gray-3'>{text}</div>}
    </div>
  )
}

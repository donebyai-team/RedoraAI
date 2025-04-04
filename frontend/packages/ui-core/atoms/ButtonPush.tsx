import { CheckIcon, CircleAlertIcon, LoaderCircleIcon } from 'lucide-react'

interface ButtonPushProps {
  onClick?: () => void
  isLoading?: boolean
  isDone?: boolean
  isError?: boolean
  disabled?: boolean
  children: React.ReactNode
  className?: string
  hideTextOnLoading?: boolean
}

export const PrimaryButtonPush = (props: ButtonPushProps) => {
  return <ButtonPush {...props} className={`w-full bg-black text-white disabled:bg-gray-3 ${props.className}`} />
}

export const ButtonPush = ({
  onClick,
  isLoading,
  isDone,
  isError,
  disabled,
  children,
  className,
  hideTextOnLoading
}: ButtonPushProps) => {
  const displayText = !isLoading || (isLoading && !hideTextOnLoading)
  return (
    <button
      className={`rounded font-semibold text-center py-1 px-2 flex flex-row gap-[10px] justify-center items-center disabled:cursor-not-allowed ${className}`}
      onClick={onClick}
      disabled={isLoading || isDone || disabled}
    >
      {isLoading ? (
        <LoaderCircleIcon size={20} className='animate-spin' />
      ) : isDone ? (
        <CheckIcon size={20} />
      ) : isError ? (
        <CircleAlertIcon size={20} />
      ) : null}
      {displayText && children}
    </button>
  )
}

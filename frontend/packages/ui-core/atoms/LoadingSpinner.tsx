import CircularProgress from '@mui/material/CircularProgress'
import { FC, useEffect, useState } from 'react'

type Props = {
  text?: string
}

export const LoadingSpinner: FC<Props> = ({ text }) => {
  const [inLoadingGracePeriod, setLoadingGracePeriod] = useState(true)

  useEffect(() => {
    const timeout = setTimeout(() => {
      setLoadingGracePeriod(false)
    }, 200)

    return () => clearTimeout(timeout)
  }, [])

  return (
    <div className='flex flex-col items-center gap-1'>
      {inLoadingGracePeriod ? null : <CircularProgress disableShrink sx={{ color: 'grey.600' }} />}
      {text && <div>{text}</div>}
    </div>
  )
}

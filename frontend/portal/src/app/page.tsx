'use client'

import { useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { routes } from '@doota/ui-core/routing'

const Home = () => {
  const { push } = useRouter()

  useEffect(() => {
    push(routes.new.dashboard)
  }, [push])

  return <p></p>
}

export default Home

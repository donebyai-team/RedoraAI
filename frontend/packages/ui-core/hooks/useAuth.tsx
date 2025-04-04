import { useContext } from 'react'
import { User } from '@doota/pb/doota/portal/v1/portal_pb'
import { AuthContext } from '../context/AuthContext'

export const useAuth = () => useContext(AuthContext)

export const useAuthUser = (): User => {
  const auth = useContext(AuthContext)
  if (!auth.user) {
    throw 'error user not defined'
  }

  return auth.user
}

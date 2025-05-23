import { User, UserRole } from "@doota/pb/doota/portal/v1/portal_pb";

export const isAdmin = (user: User): boolean => {
    return [UserRole.ADMIN, UserRole.PLATFORM_ADMIN].includes(user.role)
}

export const isPlatformAdmin = (user: User): boolean => {
    return user.role === UserRole.PLATFORM_ADMIN
}

export const isAdminUser = (user: User): boolean => {
    return UserRole.ADMIN === user.role;
}
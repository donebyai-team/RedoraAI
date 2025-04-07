'use client'

import React, { useEffect, useState } from 'react'
import { useAuthUser } from '@doota/ui-core/hooks/useAuth'
import { Box } from '@mui/system'
import {  IntegrationType } from '@doota/pb/doota/portal/v1/portal_pb'
import { FallbackSpinner } from '../../../../../atoms/FallbackSpinner'
import { Button } from '../../../../../atoms/Button'
import { portalClient } from '../../../../../services/grpc'
import { buildAppUrl } from '../../../../routes'
import { routes } from '@doota/ui-core/routing'

import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableContainer from '@mui/material/TableContainer'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Paper from '@mui/material/Paper'
import Modal from '@mui/material/Modal'
import Image from 'next/image'
import { isAdmin } from '@doota/ui-core/helper/role'

export default function Page() {
  const user = useAuthUser()
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = React.useState(false)
  const handleOpen = () => setOpen(true)
  const handleClose = () => setOpen(false)

  useEffect(() => {
  }, [setLoading])

  if (loading) {
    return <FallbackSpinner />
  }

  const openOauthConsentScreen = (integrationType: IntegrationType) => {
    portalClient
      .oauthAuthorize({
        integrationType: integrationType,
        redirectUrl: buildAppUrl(routes.app.settings.account),
      })
      .then(oAuthAuthorizeResp => {
        window.open(oAuthAuthorizeResp.authorizeUrl, '_self')
      })
  }

  return (
    <>
      {user && isAdmin(user) && (
        <Modal
          open={open}
          onClose={handleClose}
          aria-labelledby='modal-modal-title'
          aria-describedby='modal-modal-description'
        >
          <Box className='absolute top-[50%] text-center left-[50%] w-[600px] max-w-full translate-x-[-50%] translate-y-[-50%] p-2 bg-white rounded-xl'>
            <div className='text-xl font-semibold'>Select your email provider</div>
            <div className='flex mt-2 gap-2'>
              <div
                role='presentation'
                className='border border-neutral-200 hover:bg-neutral-50 hover:border-neutral-300 rounded-lg flex-1 flex items-center justify-center h-[70px] cursor-pointer'
                onClick={() => openOauthConsentScreen(IntegrationType.REDDIT)}
              >
                <Image src='/images/exchange_logo.png' alt='Exchange' width={100} height={100} />
              </div>

            </div>
          </Box>
        </Modal>
      )}

      <div className='absolute bg-neutral-50 border border-neutral-200 rounded-lg top-0 right-0 p-1.5 px-2.5 text-sm m-2.5'>
        {user && user.email}
      </div>

      <div className='flex gap-6'>
        <div className='flex-1'>
          <TableContainer
            component={Paper}
            className='inline-table border border-neutral-200 !rounded-lg !overflow-hidden !shadow-none'
          >
            {user && isAdmin(user) && (
              <>
                <div className='flex justify-end !w-full flex-1 items-end text-end p-1 border-b border-neutral-200'>
                  <Button
                    variant='contained'
                    className='!shadow-none !bg-purple-6 !text-purple-3 !rounded-md !font-semibold'
                    onClick={handleOpen}
                  >
                    Set up your provider
                  </Button>
                </div>
              </>
            )}

            <Table sx={{ minWidth: 650 }} size='small'>
              <TableHead>
                <TableRow className='bg-neutral-50'>
                  <TableCell className='!p-1'>Provider</TableCell>
                  <TableCell className='!p-1'>Mail</TableCell>
                  <TableCell align='right'></TableCell>
                </TableRow>
              </TableHead>
              <TableBody>

              </TableBody>
            </Table>
          </TableContainer>
        </div>
      </div>
    </>
  )
}

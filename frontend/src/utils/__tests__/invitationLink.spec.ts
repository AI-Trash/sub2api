import { beforeEach, describe, expect, it } from 'vitest'

import {
  INVITATION_LINK_STORAGE_KEY,
  buildInvitationLink,
  clearPersistedInvitationCode,
  extractInvitationCodeFromQuery,
  getPersistedInvitationCode,
  persistInvitationCode,
  resolveInvitationCode
} from '../invitationLink'

describe('invitationLink utils', () => {
  beforeEach(() => {
    sessionStorage.clear()
  })

  it('extracts invitation code from supported query aliases', () => {
    expect(extractInvitationCodeFromQuery({ invite: ' INVITE-1 ' })).toBe('INVITE-1')
    expect(extractInvitationCodeFromQuery({ invitation_code: 'INVITE-2' })).toBe('INVITE-2')
    expect(extractInvitationCodeFromQuery({ invitationCode: [' ', 'INVITE-3'] })).toBe('INVITE-3')
  })

  it('persists and resolves invitation codes from session storage', () => {
    expect(persistInvitationCode(' INVITE-KEEP ')).toBe('INVITE-KEEP')
    expect(sessionStorage.getItem(INVITATION_LINK_STORAGE_KEY)).toBe('INVITE-KEEP')
    expect(getPersistedInvitationCode()).toBe('INVITE-KEEP')
    expect(resolveInvitationCode({})).toBe('INVITE-KEEP')

    clearPersistedInvitationCode()

    expect(getPersistedInvitationCode()).toBe('')
    expect(sessionStorage.getItem(INVITATION_LINK_STORAGE_KEY)).toBeNull()
  })

  it('prefers query values and builds invitation links with base path support', () => {
    persistInvitationCode('OLD-CODE')

    expect(resolveInvitationCode({ invite: 'NEW-CODE' })).toBe('NEW-CODE')
    expect(getPersistedInvitationCode()).toBe('NEW-CODE')

    expect(
      buildInvitationLink('INVITE-9', {
        origin: 'https://example.com',
        basePath: '/console/',
        redirect: '/dashboard'
      })
    ).toBe('https://example.com/console/register?invite=INVITE-9&redirect=%2Fdashboard')

    expect(buildInvitationLink('INVITE-10', { basePath: '/' })).toBe(
      `${window.location.origin}/register?invite=INVITE-10`
    )
  })
})

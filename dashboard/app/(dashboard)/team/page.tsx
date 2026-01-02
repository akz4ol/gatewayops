'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Users, Mail, Shield, MoreVertical, Loader2, RefreshCw, Plus, X } from 'lucide-react';
import {
  useUsers,
  useInvites,
  useRoles,
  useCreateInvite,
  useCancelInvite,
  useResendInvite,
} from '@/lib/hooks/use-api';
import { formatDistanceToNow } from 'date-fns';
import { useSWRConfig } from 'swr';

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

function formatRelativeTime(dateString: string | undefined): string {
  if (!dateString) return 'Never';
  try {
    return formatDistanceToNow(new Date(dateString), { addSuffix: true });
  } catch {
    return 'Unknown';
  }
}

export default function TeamPage() {
  const { mutate } = useSWRConfig();
  const { data: usersData, isLoading: usersLoading, error: usersError } = useUsers();
  const { data: invitesData, isLoading: invitesLoading, error: invitesError } = useInvites();
  const { data: rolesData, isLoading: rolesLoading } = useRoles();

  const { trigger: createInvite, isMutating: isCreating } = useCreateInvite();
  const { trigger: cancelInvite } = useCancelInvite();
  const { trigger: resendInvite } = useResendInvite();

  const [showInviteForm, setShowInviteForm] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState('developer');

  const users = usersData?.users || [];
  const invites = invitesData?.invites || [];
  const roles = rolesData?.roles || [];

  const handleCreateInvite = async () => {
    if (!inviteEmail) return;
    try {
      await createInvite({ email: inviteEmail, role: inviteRole });
      setInviteEmail('');
      setInviteRole('developer');
      setShowInviteForm(false);
      mutate('invites');
    } catch (err) {
      console.error('Failed to create invite:', err);
    }
  };

  const handleCancelInvite = async (inviteId: string) => {
    try {
      await cancelInvite(inviteId);
      mutate('invites');
    } catch (err) {
      console.error('Failed to cancel invite:', err);
    }
  };

  const handleResendInvite = async (inviteId: string) => {
    try {
      await resendInvite(inviteId);
      mutate('invites');
    } catch (err) {
      console.error('Failed to resend invite:', err);
    }
  };

  const isLoading = usersLoading || invitesLoading || rolesLoading;
  const hasError = usersError || invitesError;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-600" />
      </div>
    );
  }

  if (hasError) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-4">
        <p className="text-red-600">Failed to load team data</p>
        <Button variant="outline" onClick={() => window.location.reload()}>
          <RefreshCw className="mr-2 h-4 w-4" />
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Team</h1>
          <p className="text-gray-500">Manage team members and permissions</p>
        </div>
        <Button
          className="bg-indigo-600 hover:bg-indigo-700 text-white"
          onClick={() => setShowInviteForm(true)}
        >
          <Mail className="mr-2 h-4 w-4" />
          Invite Member
        </Button>
      </div>

      {showInviteForm && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg font-medium">Invite Team Member</CardTitle>
            <Button variant="ghost" size="icon" onClick={() => setShowInviteForm(false)}>
              <X className="h-4 w-4" />
            </Button>
          </CardHeader>
          <CardContent>
            <div className="flex gap-4">
              <input
                type="email"
                placeholder="Email address"
                value={inviteEmail}
                onChange={(e) => setInviteEmail(e.target.value)}
                className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              />
              <select
                value={inviteRole}
                onChange={(e) => setInviteRole(e.target.value)}
                className="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
              >
                <option value="admin">Admin</option>
                <option value="developer">Developer</option>
                <option value="viewer">Viewer</option>
              </select>
              <Button
                onClick={handleCreateInvite}
                disabled={!inviteEmail || isCreating}
                className="bg-indigo-600 hover:bg-indigo-700 text-white"
              >
                {isCreating ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <>
                    <Plus className="mr-2 h-4 w-4" />
                    Send Invite
                  </>
                )}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-indigo-100 p-3">
                <Users className="h-5 w-5 text-indigo-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">{users.length}</p>
                <p className="text-sm text-gray-500">Team Members</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-yellow-100 p-3">
                <Mail className="h-5 w-5 text-yellow-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">{invites.length}</p>
                <p className="text-sm text-gray-500">Pending Invites</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-green-100 p-3">
                <Shield className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">{roles.length}</p>
                <p className="text-sm text-gray-500">Roles Defined</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-medium">Team Members</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="pb-3 text-left text-sm font-medium text-gray-500">Member</th>
                  <th className="pb-3 text-left text-sm font-medium text-gray-500">Role</th>
                  <th className="pb-3 text-left text-sm font-medium text-gray-500">Last Active</th>
                  <th className="pb-3 text-right text-sm font-medium text-gray-500">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {users.map((member) => (
                  <tr key={member.id} className="hover:bg-gray-50">
                    <td className="py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-indigo-100 flex items-center justify-center">
                          <span className="text-sm font-medium text-indigo-600">
                            {getInitials(member.name)}
                          </span>
                        </div>
                        <div>
                          <p className="font-medium text-gray-900">{member.name}</p>
                          <p className="text-sm text-gray-500">{member.email}</p>
                        </div>
                      </div>
                    </td>
                    <td className="py-4">
                      <span className={`rounded-full px-2 py-1 text-xs font-medium ${
                        member.role === 'admin' ? 'bg-purple-100 text-purple-700' :
                        member.role === 'developer' ? 'bg-blue-100 text-blue-700' :
                        'bg-gray-100 text-gray-700'
                      }`}>
                        {member.role.charAt(0).toUpperCase() + member.role.slice(1)}
                      </span>
                    </td>
                    <td className="py-4 text-sm text-gray-500">
                      {formatRelativeTime(member.last_active_at)}
                    </td>
                    <td className="py-4 text-right">
                      <Button variant="ghost" size="icon">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {invites.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Pending Invitations</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {invites.map((invite) => (
                <div key={invite.id} className="flex items-center justify-between rounded-lg border p-4">
                  <div>
                    <p className="font-medium text-gray-900">{invite.email}</p>
                    <p className="text-sm text-gray-500">
                      Invited by {invite.inviter_name} | {formatRelativeTime(invite.created_at)}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="rounded-full bg-blue-100 px-2 py-1 text-xs font-medium text-blue-700">
                      {invite.role.charAt(0).toUpperCase() + invite.role.slice(1)}
                    </span>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleResendInvite(invite.id)}
                    >
                      Resend
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-red-600 hover:text-red-700"
                      onClick={() => handleCancelInvite(invite.id)}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-medium">Roles</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {roles.map((role) => (
              <div key={role.id} className="rounded-lg border p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-900">
                      {role.name.charAt(0).toUpperCase() + role.name.slice(1)}
                    </p>
                    <p className="text-sm text-gray-500">{role.description}</p>
                  </div>
                </div>
                <div className="mt-2 flex flex-wrap gap-1">
                  {role.permissions.map((perm: string) => (
                    <span key={perm} className="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-600">
                      {perm}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

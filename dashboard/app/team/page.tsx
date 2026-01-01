'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Users, Mail, Shield, MoreVertical } from 'lucide-react';

const teamMembers = [
  {
    id: 1,
    name: 'Sarah Chen',
    email: 'sarah@acme.com',
    role: 'Admin',
    avatar: 'SC',
    lastActive: '2 min ago',
  },
  {
    id: 2,
    name: 'Michael Park',
    email: 'michael@acme.com',
    role: 'Developer',
    avatar: 'MP',
    lastActive: '1 hour ago',
  },
  {
    id: 3,
    name: 'Emma Wilson',
    email: 'emma@acme.com',
    role: 'Developer',
    avatar: 'EW',
    lastActive: '3 hours ago',
  },
  {
    id: 4,
    name: 'James Lee',
    email: 'james@acme.com',
    role: 'Viewer',
    avatar: 'JL',
    lastActive: '1 day ago',
  },
];

const pendingInvites = [
  {
    id: 1,
    email: 'alex@acme.com',
    role: 'Developer',
    invitedBy: 'Sarah Chen',
    sentAt: '2 days ago',
  },
];

const roles = [
  {
    name: 'Admin',
    description: 'Full access to all features and settings',
    permissions: ['*'],
  },
  {
    name: 'Developer',
    description: 'Access to MCP calls, traces, and API keys',
    permissions: ['mcp:*', 'traces:read', 'api-keys:manage'],
  },
  {
    name: 'Viewer',
    description: 'Read-only access to traces and costs',
    permissions: ['traces:read', 'costs:read'],
  },
];

export default function TeamPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Team</h1>
          <p className="text-gray-500">Manage team members and permissions</p>
        </div>
        <Button className="bg-indigo-600 hover:bg-indigo-700 text-white">
          <Mail className="mr-2 h-4 w-4" />
          Invite Member
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="rounded-full bg-indigo-100 p-3">
                <Users className="h-5 w-5 text-indigo-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">4</p>
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
                <p className="text-2xl font-bold text-gray-900">1</p>
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
                <p className="text-2xl font-bold text-gray-900">3</p>
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
                {teamMembers.map((member) => (
                  <tr key={member.id} className="hover:bg-gray-50">
                    <td className="py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-indigo-100 flex items-center justify-center">
                          <span className="text-sm font-medium text-indigo-600">{member.avatar}</span>
                        </div>
                        <div>
                          <p className="font-medium text-gray-900">{member.name}</p>
                          <p className="text-sm text-gray-500">{member.email}</p>
                        </div>
                      </div>
                    </td>
                    <td className="py-4">
                      <span className={`rounded-full px-2 py-1 text-xs font-medium ${
                        member.role === 'Admin' ? 'bg-purple-100 text-purple-700' :
                        member.role === 'Developer' ? 'bg-blue-100 text-blue-700' :
                        'bg-gray-100 text-gray-700'
                      }`}>
                        {member.role}
                      </span>
                    </td>
                    <td className="py-4 text-sm text-gray-500">{member.lastActive}</td>
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

      {pendingInvites.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium">Pending Invitations</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {pendingInvites.map((invite) => (
                <div key={invite.id} className="flex items-center justify-between rounded-lg border p-4">
                  <div>
                    <p className="font-medium text-gray-900">{invite.email}</p>
                    <p className="text-sm text-gray-500">
                      Invited by {invite.invitedBy} | {invite.sentAt}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="rounded-full bg-blue-100 px-2 py-1 text-xs font-medium text-blue-700">
                      {invite.role}
                    </span>
                    <Button variant="outline" size="sm">Resend</Button>
                    <Button variant="ghost" size="sm" className="text-red-600 hover:text-red-700">Cancel</Button>
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
              <div key={role.name} className="rounded-lg border p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium text-gray-900">{role.name}</p>
                    <p className="text-sm text-gray-500">{role.description}</p>
                  </div>
                </div>
                <div className="mt-2 flex flex-wrap gap-1">
                  {role.permissions.map((perm) => (
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

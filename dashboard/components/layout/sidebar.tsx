'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  Activity,
  DollarSign,
  Key,
  Settings,
  Shield,
  Bell,
  Users,
  FileText,
  CheckSquare,
  Bot,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useAuth } from '@/lib/auth';

const navigation = [
  { name: 'Overview', href: '/', icon: LayoutDashboard },
  { name: 'Traces', href: '/traces', icon: Activity },
  { name: 'Costs', href: '/costs', icon: DollarSign },
  { name: 'API Keys', href: '/api-keys', icon: Key },
  { name: 'Agents', href: '/agents', icon: Bot },
  { name: 'Safety', href: '/safety', icon: Shield },
  { name: 'Alerts', href: '/alerts', icon: Bell },
  { name: 'Team', href: '/team', icon: Users },
  { name: 'Settings', href: '/settings', icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user } = useAuth();

  const orgName = user?.org_name || 'Organization';
  const orgInitial = orgName.charAt(0).toUpperCase();

  return (
    <div className="flex h-full w-64 flex-col bg-gray-900">
      <div className="flex h-16 items-center gap-2 px-6">
        <div className="h-8 w-8 rounded-lg bg-indigo-500 flex items-center justify-center">
          <span className="text-white font-bold text-lg">G</span>
        </div>
        <span className="text-xl font-semibold text-white">GatewayOps</span>
      </div>

      <nav className="flex-1 space-y-1 px-3 py-4">
        {navigation.map((item) => {
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-gray-800 text-white'
                  : 'text-gray-400 hover:bg-gray-800 hover:text-white'
              )}
            >
              <item.icon className="h-5 w-5" />
              {item.name}
            </Link>
          );
        })}
      </nav>

      <div className="border-t border-gray-800 p-4">
        <div className="flex items-center gap-3">
          <div className="h-8 w-8 rounded-full bg-blue-600 flex items-center justify-center">
            <span className="text-sm font-medium text-white">{orgInitial}</span>
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-white truncate">{orgName}</p>
            <p className="text-xs text-gray-400 truncate">Production</p>
          </div>
        </div>
      </div>
    </div>
  );
}

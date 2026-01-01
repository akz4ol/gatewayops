'use client';

import { LucideIcon } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface StatsCardProps {
  title: string;
  value: string;
  change: number;
  icon: LucideIcon;
  description: string;
  changeType?: 'positive' | 'negative';
}

export function StatsCard({
  title,
  value,
  change,
  icon: Icon,
  description,
  changeType,
}: StatsCardProps) {
  const isPositive = changeType
    ? changeType === 'positive'
    : change >= 0;

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-gray-500">{title}</p>
            <p className="text-2xl font-bold text-gray-900">{value}</p>
          </div>
          <div className="rounded-full bg-gray-100 p-3">
            <Icon className="h-5 w-5 text-gray-600" />
          </div>
        </div>
        <div className="mt-4 flex items-center gap-2">
          <span
            className={cn(
              'text-sm font-medium',
              isPositive ? 'text-green-600' : 'text-red-600'
            )}
          >
            {isPositive ? '+' : ''}
            {change}%
          </span>
          <span className="text-sm text-gray-500">{description}</span>
        </div>
      </CardContent>
    </Card>
  );
}

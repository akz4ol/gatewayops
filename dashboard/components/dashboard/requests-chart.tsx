'use client';

import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';

// Sample data - in production this would come from the API
const data = [
  { date: 'Dec 1', requests: 4000, errors: 24 },
  { date: 'Dec 2', requests: 3000, errors: 13 },
  { date: 'Dec 3', requests: 5000, errors: 22 },
  { date: 'Dec 4', requests: 4500, errors: 18 },
  { date: 'Dec 5', requests: 6000, errors: 29 },
  { date: 'Dec 6', requests: 5500, errors: 25 },
  { date: 'Dec 7', requests: 7000, errors: 32 },
  { date: 'Dec 8', requests: 6500, errors: 28 },
  { date: 'Dec 9', requests: 8000, errors: 35 },
  { date: 'Dec 10', requests: 7500, errors: 30 },
  { date: 'Dec 11', requests: 9000, errors: 42 },
  { date: 'Dec 12', requests: 8500, errors: 38 },
  { date: 'Dec 13', requests: 10000, errors: 45 },
  { date: 'Dec 14', requests: 9500, errors: 40 },
];

export function RequestsChart() {
  return (
    <div className="h-[300px]">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
          <defs>
            <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
              <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
          <XAxis
            dataKey="date"
            axisLine={false}
            tickLine={false}
            tick={{ fill: '#9ca3af', fontSize: 12 }}
          />
          <YAxis
            axisLine={false}
            tickLine={false}
            tick={{ fill: '#9ca3af', fontSize: 12 }}
            tickFormatter={(value) => `${value / 1000}k`}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: '#fff',
              border: '1px solid #e5e7eb',
              borderRadius: '8px',
              boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)',
            }}
            formatter={(value: number) => [value.toLocaleString(), 'Requests']}
          />
          <Area
            type="monotone"
            dataKey="requests"
            stroke="#6366f1"
            strokeWidth={2}
            fillOpacity={1}
            fill="url(#colorRequests)"
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: 'standalone',
  experimental: {
    serverActions: true,
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.GATEWAYOPS_API_URL || 'http://localhost:8080'}/v1/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;

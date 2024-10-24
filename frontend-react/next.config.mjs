/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: false,
  async redirects() {
    return [
        {
            source: '/',
            destination: '/dashboard', // Redirect root to dashboard
            permanent: true, // Permanent redirect (301)
        },
    ];
  },
};

export default nextConfig;

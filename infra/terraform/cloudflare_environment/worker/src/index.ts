/**
 * Welcome to Cloudflare Workers! This is your first worker.
 * Learn more at https://developers.cloudflare.com/workers/
 */

declare namespace Cloudflare {
  // eslint-disable-next-line @typescript-eslint/no-empty-interface,@typescript-eslint/no-empty-object-type
  interface Env {}
}
interface Env extends Cloudflare.Env {}

export default {
  async fetch(request, env, ctx): Promise<Response> {
    const url = new URL(request.url);

    // Handle both the root path and any URL with query parameters
    if (url.pathname === '/' || url.pathname === '') {
      // Create a new URL with the same origin but /index.html path
      // Preserve any query parameters from the original URL
      const indexUrl = new URL('/index.html', url.origin);

      // Copy all query parameters from the original URL
      url.searchParams.forEach((value, key) => {
        indexUrl.searchParams.append(key, value);
      });

      // Return a redirect response with caching headers
      return new Response(null, {
        status: 301,
        headers: {
          Location: indexUrl.toString(),
          'Cache-Control': 'public, max-age=86400',
          'CDN-Cache-Control': 'max-age=86400',
        },
      });
    }

    // This code should never be reached if routes are configured correctly
    // But just in case, pass through to the origin to avoid 404 errors
    return fetch(request);
  },
} satisfies ExportedHandler<Env>;

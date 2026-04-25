import type { Handle, HandleFetch } from '@sveltejs/kit';

// Attach a correlation ID to every request.
export const handle: Handle = async ({ event, resolve }) => {
  const reqId = event.request.headers.get('x-request-id') ?? crypto.randomUUID();
  event.locals.tenantId = event.request.headers.get('x-tenant-id') ?? undefined;
  event.locals.userEmail = event.request.headers.get('x-user-email') ?? undefined;
  const response = await resolve(event, {
    filterSerializedResponseHeaders: (name) => ['content-type', 'content-length'].includes(name.toLowerCase())
  });
  response.headers.set('x-request-id', reqId);
  response.headers.set('x-frame-options', 'DENY');
  response.headers.set('x-content-type-options', 'nosniff');
  response.headers.set('referrer-policy', 'strict-origin-when-cross-origin');
  response.headers.set('permissions-policy', 'geolocation=(), camera=(), microphone=()');
  response.headers.set(
    'content-security-policy',
    "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self' https://api.terna.it https://api.e-distribuzione.it"
  );
  return response;
};

// Forward the request's correlation ID to backend fetches.
export const handleFetch: HandleFetch = async ({ event, request, fetch }) => {
  const reqId = event.request.headers.get('x-request-id');
  if (reqId) {
    request.headers.set('x-request-id', reqId);
  }
  return fetch(request);
};

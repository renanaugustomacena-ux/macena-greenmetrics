// See https://kit.svelte.dev/docs/types#app
declare global {
  namespace App {
    interface Error {
      code?: string;
      requestId?: string;
    }
    interface Locals {
      userEmail?: string;
      tenantId?: string;
      role?: 'admin' | 'manager' | 'operator' | 'auditor' | 'readonly';
    }
    interface PageData {}
    interface PageState {}
    interface Platform {}
  }
}

export {};

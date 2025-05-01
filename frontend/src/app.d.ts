// See https://svelte.dev/docs/kit/types#app.d.ts

import type { Session, SupabaseClient, User } from "@supabase/supabase-js";

// for information about these interfaces
declare global {
	namespace App {
		// interface Locals {}
		// interface PageData {}
		// interface Error {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};

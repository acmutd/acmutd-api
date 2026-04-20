import { initializeApp } from "firebase/app";
import { getAuth } from "firebase/auth";

// The Go backend injects this global by replacing the "__FIREBASE_CONFIG__" placeholder
// in index.html with the real config JSON object at request time (see dashboard_pages.go).
declare const __FIREBASE_CONFIG__: { apiKey: string; authDomain: string; projectId: string };

export const firebaseApp = initializeApp(__FIREBASE_CONFIG__);
export const firebaseAuth = getAuth(firebaseApp);

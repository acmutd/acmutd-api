import { initializeApp } from "firebase/app";
import { getAuth } from "firebase/auth";

// In production the Go server injects the config by replacing __FIREBASE_CONFIG__
// in the HTML with a real JSON object. In local dev (npm run dev) we fall back
// to Vite environment variables defined in .env.
declare const __FIREBASE_CONFIG__: string | undefined;

const config =
  typeof __FIREBASE_CONFIG__ !== "undefined" && __FIREBASE_CONFIG__ !== "{}"
    ? (JSON.parse(__FIREBASE_CONFIG__) as object)
    : {
        apiKey: import.meta.env.VITE_FIREBASE_API_KEY as string,
        authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN as string,
        projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID as string,
      };

export const firebaseApp = initializeApp(config);
export const firebaseAuth = getAuth(firebaseApp);

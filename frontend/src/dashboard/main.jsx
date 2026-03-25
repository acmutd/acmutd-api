import React from "react";
import { createRoot } from "react-dom/client";
import { DashboardApp } from "./DashboardApp";
import "./tailwind.css";

createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <DashboardApp />
  </React.StrictMode>
);

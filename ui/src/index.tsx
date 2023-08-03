import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import {
  createGlobalFadeTransition,
  ScreenSizeProvider,
  TransitionDuration,
} from "./components"
import { ThemeProvider } from '@primer/react'
import { LocalStorageProvider } from "./providers/LocalStorageProvider"


const FadeReg = createGlobalFadeTransition("fade-reg", TransitionDuration.REG)

const FadeSlow = createGlobalFadeTransition(
  "fade-slow",
  TransitionDuration.SLOW,
)
const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

root.render(
  <ScreenSizeProvider>
    <ThemeProvider>
      <LocalStorageProvider>
        <FadeSlow />
        <FadeReg />
        <App />
      </LocalStorageProvider>
    </ThemeProvider>
  </ScreenSizeProvider>,
)

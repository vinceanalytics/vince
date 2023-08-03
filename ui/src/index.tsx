import React from 'react';
import ReactDOM from 'react-dom/client';
import {
  createGlobalFadeTransition,
  ScreenSizeProvider,
  TransitionDuration,
} from "./components"
import { ThemeProvider } from '@primer/react'
import { LocalStorageProvider } from "./providers/LocalStorageProvider"
import Layout from "./scenes/Layout"


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
        <Layout />
      </LocalStorageProvider>
    </ThemeProvider>
  </ScreenSizeProvider>,
)

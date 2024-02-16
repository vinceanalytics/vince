import React from 'react';
import { ThemeProvider, BaseStyles, Box, Heading } from '@primer/react'
import { Dashboard } from "./pages";
import { VinceProvider } from "./providers";
function App() {
  return (
    <ThemeProvider>
      <BaseStyles>
        <VinceProvider>
          <Dashboard />
        </VinceProvider>
      </BaseStyles>
    </ThemeProvider>
  );
}

export default App;

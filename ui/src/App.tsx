import React from 'react';
import { ThemeProvider, BaseStyles, Box, Heading } from '@primer/react'
import { Dashboard } from "./pages";
function App() {
  return (
    <ThemeProvider>
      <BaseStyles>
        <Dashboard />
      </BaseStyles>
    </ThemeProvider>
  );
}

export default App;

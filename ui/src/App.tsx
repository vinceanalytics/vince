import React from 'react';
import { ThemeProvider, BaseStyles, Box, Heading } from '@primer/react'

function App() {
  return (
    <ThemeProvider>
      <BaseStyles>
        <Box m={4}>
          <Heading as="h2" sx={{ mb: 2 }}>
            Hello, world!
          </Heading>
          <p>This will get Primer text styles.</p>
        </Box>
      </BaseStyles>
    </ThemeProvider>
  );
}

export default App;

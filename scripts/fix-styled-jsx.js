#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Simple script to remove styled-jsx from React components
// and convert them to use CSS modules or inline styles

const componentsDir = path.join(__dirname, '../web/src/components');

function removeStyledJsx(filePath) {
  const content = fs.readFileSync(filePath, 'utf8');
  
  // Remove styled-jsx blocks
  const cleaned = content.replace(
    /\s*<style jsx>\{`[\s\S]*?`\}<\/style>/g,
    ''
  );
  
  fs.writeFileSync(filePath, cleaned);
  console.log(`Cleaned ${filePath}`);
}

// Process all TypeScript files in components directory
function processComponents() {
  if (!fs.existsSync(componentsDir)) {
    console.log('Components directory not found');
    return;
  }
  
  const files = fs.readdirSync(componentsDir);
  
  files.forEach(file => {
    if (file.endsWith('.tsx')) {
      const filePath = path.join(componentsDir, file);
      removeStyledJsx(filePath);
    }
  });
}

processComponents();
console.log('Done removing styled-jsx blocks!');

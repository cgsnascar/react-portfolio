import React from 'react';
import Header from './components/header';
import About from './pages/About';
import Projects from './pages/Projects';
import Contact from './pages/Contact';
import './styles/App.css';
import DataDisplay from './components/DataDisplay';

function App() {
  return (
    <div className="App">
      <Header />
      <About />
      <Projects />
      <DataDisplay />
      <Contact />
    </div>
  );
}

export default App;

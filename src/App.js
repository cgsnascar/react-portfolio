import React from 'react';
import { Helmet } from 'react-helmet';
import Header from './components/header';
import About from './pages/About';
import Projects from './pages/Projects';
import Contact from './pages/Contact';
import './styles/App.css';
import DataDisplay from './components/DataDisplay';

function App() {
  return (
    <div className="App">
      <Helmet>
        <title>Claudio Skala</title>
        <meta name="description" content="Website portfolio for Claudio Skala" />
        <meta property="og:title" content="Claudio Skala" />
        <meta property="og:description" content="Website portfolio for Claudio Skala" />
      </Helmet>
      <Header />
      <About />
      <Projects />
      <DataDisplay />
      <Contact />
    </div>
  );
}

export default App;

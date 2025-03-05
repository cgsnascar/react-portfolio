import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { Helmet } from "react-helmet";
import Header from "./components/header";
import About from "./pages/About";
import Projects from "./pages/Projects";
import Contact from "./pages/Contact";
import Footer from "./components/Footer";
import "./styles/App.css";
import DataDisplay from "./components/DataDisplay";

function App() {
  return (
    <Router>
      <div className="App">
        <Helmet>
          <title>Claudio Skala | Junior Web Developer</title>
          <meta
            name="description"
            content="Claudio Skala is an aspiring junior web developer."
          />
          <meta property="og:title" content="Claudio Skala Web Portfolio" />
          <meta
            property="og:description"
            content="Claudio Skala is an aspiring junior web developer."
          />
        </Helmet>

        <Header />

        <Routes>
          <Route
            path="/"
            element={
              <>
                <About />
                <Projects />
                <DataDisplay />
                <Contact />
              </>
            }
          />
        </Routes>

        <Footer />
      </div>
    </Router>
  );
}

export default App;

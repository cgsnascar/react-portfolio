import React from 'react';

function Header() {
  return (
    <header className="bg-gray-800 text-white p-4">
      <nav>
        <ul className="flex justify-center space-x-4">
          <li><a href="#about" className="hover:text-yellow-500">About</a></li>
          <li><a href="#projects" className="hover:text-yellow-500">Projects</a></li>
          <li><a href="#contact" className="hover:text-yellow-500">Contact</a></li>
        </ul>
      </nav>
    </header>
  );
}

export default Header;

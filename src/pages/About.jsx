import React from 'react';

function About() {
  return (
    <section id="about" className="p-8 text-center bg-gray-100">
      <h2 className="text-4xl font-bold mb-4">About Me</h2>
      <p className="text-lg">Hello! My name is Claudio, I am learning about frontend and backend web development.</p>
      <p>This website was created using React.JS for the front end, and the back end consists on multiple API created using GO</p>
      <p className='max-w-2xl mx-auto p-2'>While I am still learning frontend and some backend development, I will be updating this website with the projects I am working on, and any reviews I may get from new or previous clients. This whole website is still a work in progress, but I am having fun learning along the way.</p>
      <p className='max-w-2xl mx-auto p-2'>I consider myself a junior frontend developer who likes to dabble in backend development on occassion. By current schooling, I am an IT Professional, but I am trying to follow my passion for web development. and this is my first big step into this. I will try to keep my projects updated as I continue developing this website.</p>
      <p className='max-w-2xl mx-auto p-2'>I will be updating quite a bit of things on the backend soon, including an admin panel of some kind to make the about section, projects, and accepting reviews possible where a key isn't required, but something I could just approve. I intend to adjust this section quickly and put all of this nonsense into a blog-like section!</p>
      <p className='max-w-2xl mx-auto p-2'>If you notice any issues with this site, or something not loading correctly please feel free to reach out via the contact form below, if that does not work, please email me <a href='mailto:cskala@cgsnascar.dev'>here</a>, and I will get back to you as soon as possible.</p>
    </section>
  );
}

export default About;

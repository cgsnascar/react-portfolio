import React, { useState, useEffect } from 'react';

function Projects() {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetch('http://localhost:8080/api/projects')
      .then(response => response.json())
      .then(data => {
        setProjects(Array.isArray(data) ? data : []);
        setLoading(false);
      })
      .catch(error => {
        console.error('Error fetching projects:', error);
        setError(error);
        setLoading(false);
      });
  }, []);

  if (loading) {
    return <p className="text-center text-gray-600">Loading...</p>;
  }

  if (error) {
    return <p className="text-center text-red-600">Error loading projects.</p>;
  }

  return (
    <section id="projects" className="bg-gray-100 py-8">
      <div className="container mx-auto px-4">
        <h1 className="text-3xl font-bold text-center mb-8">My Projects</h1>
        {projects.length === 0 ? (
          <div className="flex justify-center items-center h-48">
            <p className="text-center text-gray-600 text-xl">No projects available</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {projects.map(project => (
              <div key={project.id} className="bg-white shadow-lg rounded-lg overflow-hidden">
                <div className="p-6">
                  <h2 className="text-xl font-semibold mb-2">{project.title}</h2>
                  <p className="text-gray-700 mb-4">{project.description}</p>
                  <a 
                    href={project.url} 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="text-indigo-600 hover:text-indigo-800 font-semibold"
                  >
                    {project.actionLabel} {/* Dynamic label: Show Code or Show Website */}
                  </a>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}

export default Projects;

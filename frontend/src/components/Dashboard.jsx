import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { LogOut, Plus, Trello, CalendarDays, LayoutDashboard, Moon, Sun, Trash2, ListFilter } from 'lucide-react';
import api from '../api';
import Board from './Board';
import TaskModal from './TaskModal';
import { Spinner } from './Loader';

const Dashboard = ({ setAuth }) => {
  const [theme, setTheme] = useState(localStorage.getItem('theme') || 'dark');
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false);
  const [taskToEdit, setTaskToEdit] = useState(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [projects, setProjects] = useState([]);
  const [activeProjectIds, setActiveProjectIds] = useState([]);
  const [newProjectName, setNewProjectName] = useState('');
  const [isCreatingProject, setIsCreatingProject] = useState(false);
  const [isFetchingProjects, setIsFetchingProjects] = useState(false);
  const [isCreatingProjectLoading, setIsCreatingProjectLoading] = useState(false);
  const [viewMode, setViewMode] = useState('weekly'); // 'weekly', 'global', 'backlog', 'project'
  const navigate = useNavigate();

  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = async () => {
    setIsFetchingProjects(true);
    try {
      const res = await api.get('/projects');
      setProjects(res.data || []);
    } catch (err) {
      console.error('Failed to fetch projects', err);
      if (err.response?.status === 401) handleLogout();
    } finally {
      setIsFetchingProjects(false);
    }
  };

  const handleCreateProject = async (e) => {
    e.preventDefault();
    if (!newProjectName.trim()) return;
    setIsCreatingProjectLoading(true);
    try {
      const res = await api.post('/projects', { name: newProjectName });
      setProjects([...projects, res.data]);
      setActiveProjectIds([res.data.id]);
      setNewProjectName('');
      setIsCreatingProject(false);
      setViewMode('project');
    } catch (err) {
      console.error('Failed to create project', err);
    } finally {
      setIsCreatingProjectLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setAuth(false);
    navigate('/auth');
  };

  const toggleTheme = async () => {
    const newTheme = theme === 'dark' ? 'light' : 'dark';
    setTheme(newTheme);
    localStorage.setItem('theme', newTheme);
    document.documentElement.setAttribute('data-theme', newTheme);
    try {
      await api.patch('/user/theme', { theme: newTheme });
    } catch(err) {
      console.error(err);
    }
  };

  const handleOpenCreateModal = () => {
    setTaskToEdit(null);
    setIsTaskModalOpen(true);
  };
  
  const handleOpenEditModal = (task) => {
    setTaskToEdit(task);
    setIsTaskModalOpen(true);
  };

  const handleTaskSave = (savedTask) => {
    setRefreshTrigger(prev => prev + 1);
  };

  const toggleProject = (id) => {
    setActiveProjectIds(prev => {
      const isSelected = prev.includes(id);
      const newIds = isSelected ? prev.filter(pId => pId !== id) : [...prev, id];
      if (newIds.length === 0) {
        setViewMode('weekly'); // Default to weekly if no projects selected
      } else {
        setViewMode('project');
      }
      return newIds;
    });
  };

  const handleDeleteProject = async (e, id, name) => {
    e.stopPropagation(); // Prevent toggling project
    if (window.confirm(`Warning: Deleting the project "${name}" will permanently delete ALL tasks inside it. Are you sure?`)) {
      try {
        await api.delete(`/projects/${id}`);
        setProjects(projects.filter(p => p.id !== id));
        if (activeProjectIds.includes(id)) {
          toggleProject(id); // Removes from activeProjectIds
        }
        setRefreshTrigger(prev => prev + 1);
      } catch (err) {
        console.error('Failed to delete project', err);
      }
    }
  };

  const getNavTitle = () => {
    if (viewMode === 'weekly') return 'Weekly Board';
    if (viewMode === 'backlog') return 'Backlog Tasks';
    if (viewMode === 'global') return 'Global Kanban Board';
    if (viewMode === 'project') {
      if (activeProjectIds.length === 1) {
        return projects.find(p => p.id === activeProjectIds[0])?.name || 'Project Board';
      }
      return 'Multiple Projects';
    }
    return 'Kanban Board';
  };

  return (
    <div className="dashboard-container">
      {/* Sidebar */}
      <aside className="sidebar">
        <div className="sidebar-header">
          <h2>Karma</h2>
        </div>
        
        <div className="sidebar-menu">
          <button 
            className={`menu-btn ${viewMode === 'weekly' ? 'active' : ''}`}
            onClick={() => {
              setActiveProjectIds([]);
              setViewMode('weekly');
            }}
          >
            <CalendarDays size={18} /> Weekly Board
          </button>
          <button 
            className={`menu-btn ${viewMode === 'global' ? 'active' : ''}`}
            onClick={() => {
              setActiveProjectIds([]);
              setViewMode('global');
            }}
          >
            <LayoutDashboard size={18} /> Global Board
          </button>
          <button 
            className={`menu-btn ${viewMode === 'backlog' ? 'active' : ''}`}
            onClick={() => {
              setActiveProjectIds([]);
              setViewMode('backlog');
            }}
          >
            <ListFilter size={18} /> Backlog Tasks
          </button>
        </div>

        <ul className="project-list">
          <li className="list-heading" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            FILTER BY PROJECT {isFetchingProjects && <Spinner size={12} />}
          </li>
          {projects.map(p => (
            <li 
              key={p.id} 
              className={`project-item ${viewMode === 'project' && activeProjectIds.includes(p.id) ? 'active' : ''}`}
              style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}
              onClick={() => toggleProject(p.id)}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Trello size={16} /> {p.name}
              </div>
              <button className="icon-btn text-btn" onClick={(e) => handleDeleteProject(e, p.id, p.name)} style={{ padding: '0.1rem' }}>
                <Trash2 size={14} color="var(--danger)" />
              </button>
            </li>
          ))}
          
          {isCreatingProject ? (
            <form onSubmit={handleCreateProject} style={{ padding: '0 1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <input 
                type="text" 
                value={newProjectName} 
                onChange={(e) => setNewProjectName(e.target.value)} 
                placeholder="Project Name..."
                autoFocus
                onBlur={() => {if(!newProjectName) setIsCreatingProject(false);}}
                className="inline-input"
                disabled={isCreatingProjectLoading}
              />
              {isCreatingProjectLoading && <Spinner size={14} />}
            </form>
          ) : (
            <li className="project-item create-btn" onClick={() => setIsCreatingProject(true)}>
              <Plus size={16} /> New Project
            </li>
          )}
        </ul>
      </aside>

      {/* Main Content */}
      <main className="main-content">
        <header className="top-nav">
          <div className="nav-title" style={{ fontWeight: 'bold' }}>
            {getNavTitle()}
          </div>
          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
            <button className="primary-btn btn-small flex-center" onClick={handleOpenCreateModal}>
              <Plus size={16} /> Create Task
            </button>
            <button onClick={toggleTheme} className="text-btn flex-center" style={{ padding: '0.25rem', borderRadius: '50%', background: 'var(--bg-card)', border: '1px solid var(--border-color)' }}>
              {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
            </button>
            <button onClick={handleLogout} className="text-btn flex-center">
              <LogOut size={16} /> Logout
            </button>
          </div>
        </header>

        <Board 
          viewMode={viewMode} 
          activeProjectIds={activeProjectIds} 
          projects={projects} 
          refreshTrigger={refreshTrigger} 
          onEditTask={handleOpenEditModal} 
        />
      </main>
      
      <TaskModal 
        isOpen={isTaskModalOpen}
        onClose={() => setIsTaskModalOpen(false)}
        task={taskToEdit}
        projects={projects}
        onSave={handleTaskSave}
      />
    </div>
  );
};

export default Dashboard;

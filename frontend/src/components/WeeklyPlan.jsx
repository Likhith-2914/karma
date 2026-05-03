import React, { useState, useEffect } from 'react';
import { Calendar, AlertCircle, Edit2, Folder } from 'lucide-react';
import api from '../api';

const PROJECT_COLORS = [
  '#0052CC', // Jira blue
  '#00875A', // Green
  '#FF5630', // Red
  '#FFAB00', // Yellow
  '#36B37E', // Light green
  '#00B8D9', // Cyan
  '#6554C0', // Purple
];

const WeeklyPlan = ({ refreshTrigger, onEditTask, projects }) => {
  const [tasks, setTasks] = useState([]);
  
  useEffect(() => {
    fetchTasks();
  }, [refreshTrigger]);

  const fetchTasks = async () => {
    try {
      const res = await api.get('/tasks');
      setTasks(res.data || []);
    } catch (err) {
      console.error('Failed to fetch tasks', err);
    }
  };

  // Helper function to get start and end of current week (Monday - Sunday)
  const getWeekDates = () => {
    const today = new Date();
    const day = today.getDay(); // 0 is Sunday, 1 is Monday...
    
    // Calculate Monday
    const diffToMonday = today.getDate() - day + (day === 0 ? -6 : 1);
    const startOfWeek = new Date(today.setDate(diffToMonday));
    startOfWeek.setHours(0, 0, 0, 0);

    // Calculate Sunday
    const endOfWeek = new Date(startOfWeek);
    endOfWeek.setDate(startOfWeek.getDate() + 6);
    endOfWeek.setHours(23, 59, 59, 999);

    return { startOfWeek, endOfWeek };
  };

  const { startOfWeek, endOfWeek } = getWeekDates();

  // Categorize Tasks
  const overdueTasks = [];
  const thisWeekTasks = [];
  const upcomingTasks = [];

  tasks.forEach(task => {
    if (task.status === 'COMPLETED' || task.status === 'CANCELLED') return; // Only show active tasks
    if (!task.due_date) return; // Skip tasks without dates for the planner

    const taskDate = new Date(task.due_date);

    if (taskDate < startOfWeek) {
      overdueTasks.push(task); // Missed tasks carried forward
    } else if (taskDate >= startOfWeek && taskDate <= endOfWeek) {
      thisWeekTasks.push(task);
    } else {
      upcomingTasks.push(task);
    }
  });

  const getProjectName = (id) => {
    if (!id || id === 0) return 'UNASSIGNED';
    const project = projects.find(p => p.id === id);
    return project ? project.name : 'UNASSIGNED';
  };

  const getProjectColor = (id) => {
    if (!id || id === 0) return '#6b778c'; // Unassigned grey
    return PROJECT_COLORS[id % PROJECT_COLORS.length];
  };

  return (
    <div className="weekly-plan-container">
      <div className="weekly-header">
        <h2>Your Weekly Plan</h2>
        <p className="subtitle">Focus on what matters this week. Missed tasks automatically carry forward.</p>
      </div>

      <div className="plan-sections">
        {overdueTasks.length > 0 && (
          <section className="plan-section overdue-section">
            <h3>
              <AlertCircle size={20} color="var(--danger)" /> 
              Carried Forward (Overdue)
            </h3>
            <div className="task-list-simple">
              {overdueTasks.map(t => {
                const projectName = getProjectName(t.project_id);
                return (
                <div key={t.id} className="simple-task-card overdue">
                  <div style={{ display: 'flex', flexDirection: 'column' }}>
                    <span className="task-title">{t.title}</span>
                    <div style={{ display: 'flex', alignItems: 'center' }}>
                      <span className="task-date">Due: {t.due_date}</span>
                      <span className="task-project-badge" style={{ color: '#ffffff', backgroundColor: getProjectColor(t.project_id), marginLeft: '0.5rem', marginBottom: 0 }}>
                        <Folder size={8} /> {projectName}
                      </span>
                    </div>
                  </div>
                  <button className="icon-btn text-btn" onClick={() => onEditTask(t)}><Edit2 size={14}/></button>
                </div>
              )})}
            </div>
          </section>
        )}

        <section className="plan-section">
          <h3>
            <Calendar size={20} color="var(--accent-primary)" /> 
            This Week ({startOfWeek.toLocaleDateString()} - {endOfWeek.toLocaleDateString()})
          </h3>
          {thisWeekTasks.length === 0 ? (
            <p className="empty-state">No active tasks scheduled for this week. Use the Kanban board to create some!</p>
          ) : (
            <div className="task-list-simple">
              {thisWeekTasks.map(t => {
                const projectName = getProjectName(t.project_id);
                return (
                <div key={t.id} className="simple-task-card">
                  <div style={{ display: 'flex', flexDirection: 'column' }}>
                    <span className="task-title">{t.title}</span>
                    <div style={{ display: 'flex', alignItems: 'center' }}>
                      <span className="task-date">Due: {t.due_date}</span>
                      <span className="task-project-badge" style={{ color: '#ffffff', backgroundColor: getProjectColor(t.project_id), marginLeft: '0.5rem', marginBottom: 0 }}>
                        <Folder size={8} /> {projectName}
                      </span>
                    </div>
                  </div>
                  <button className="icon-btn text-btn" onClick={() => onEditTask(t)}><Edit2 size={14}/></button>
                </div>
              )})}
            </div>
          )}
        </section>

        {upcomingTasks.length > 0 && (
          <section className="plan-section">
            <h3>Upcoming (Later)</h3>
            <div className="task-list-simple">
              {upcomingTasks.map(t => {
                const projectName = getProjectName(t.project_id);
                return (
                <div key={t.id} className="simple-task-card upcoming">
                  <div style={{ display: 'flex', flexDirection: 'column' }}>
                    <span className="task-title">{t.title}</span>
                    <div style={{ display: 'flex', alignItems: 'center' }}>
                      <span className="task-date">Due: {t.due_date}</span>
                      <span className="task-project-badge" style={{ color: '#ffffff', backgroundColor: getProjectColor(t.project_id), marginLeft: '0.5rem', marginBottom: 0 }}>
                        <Folder size={8} /> {projectName}
                      </span>
                    </div>
                  </div>
                  <button className="icon-btn text-btn" onClick={() => onEditTask(t)}><Edit2 size={14}/></button>
                </div>
              )})}
            </div>
          </section>
        )}
      </div>
    </div>
  );
};

export default WeeklyPlan;

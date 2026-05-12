import React, { useState, useEffect } from 'react';
import { DragDropContext, Droppable, Draggable } from '@hello-pangea/dnd';
import { Plus, Calendar as CalIcon, Folder, Edit2 } from 'lucide-react';
import api from '../api';
import { Spinner } from './Loader';

const COLUMNS = [
  { id: 'TODO', title: 'To Do' },
  { id: 'IN_PROGRESS', title: 'In Progress' },
  { id: 'COMPLETED', title: 'Completed' },
  { id: 'CANCELLED', title: 'Cancelled' }
];

const PROJECT_COLORS = [
  '#0052CC', // Jira blue
  '#00875A', // Green
  '#FF5630', // Red
  '#FFAB00', // Yellow
  '#36B37E', // Light green
  '#00B8D9', // Cyan
  '#6554C0', // Purple
];

const Board = ({ viewMode, activeProjectIds = [], projects, refreshTrigger, onEditTask }) => {
  const [tasks, setTasks] = useState([]);
  const [isFetchingTasks, setIsFetchingTasks] = useState(false);

  useEffect(() => {
    fetchTasks();
  }, [refreshTrigger]);

  const fetchTasks = async () => {
    setIsFetchingTasks(true);
    try {
      const res = await api.get('/tasks');
      setTasks(res.data || []);
    } catch (err) {
      console.error('Failed to fetch tasks', err);
    } finally {
      setIsFetchingTasks(false);
    }
  };



  const onDragEnd = async (result) => {
    const { destination, source, draggableId } = result;
    if (!destination) return;
    if (destination.droppableId === source.droppableId && destination.index === source.index) return;

    const newStatus = destination.droppableId;
    const taskToMove = tasks.find(t => t.id.toString() === draggableId);
    if (!taskToMove) return;

    const updatedTasks = tasks.map(t => 
      t.id === taskToMove.id ? { ...t, status: newStatus } : t
    );
    setTasks(updatedTasks);

    try {
      await api.patch(`/tasks/${taskToMove.id}/status`, { status: newStatus });
    } catch (err) {
      console.error('Failed to update task status', err);
      fetchTasks();
    }
  };

  const isOverdue = (dateStr) => {
    if (!dateStr) return false;
    const today = new Date();
    today.setHours(0,0,0,0);
    const taskDate = new Date(dateStr);
    return taskDate < today;
  };

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

  const getProjectName = (id) => {
    if (!id || id === 0) return 'UNASSIGNED';
    const project = projects.find(p => p.id === id);
    return project ? project.name : 'UNASSIGNED';
  };

  const getProjectColor = (id) => {
    if (!id || id === 0) return '#6b778c'; // Unassigned grey
    return PROJECT_COLORS[id % PROJECT_COLORS.length];
  };

  let visibleTasks = tasks;

  if (viewMode === 'project') {
    visibleTasks = tasks.filter(t => activeProjectIds.includes(t.project_id));
  } else if (viewMode === 'weekly') {
    visibleTasks = tasks.filter(t => {
      if (!t.due_date) return false;
      const d = new Date(t.due_date);
      const isThisWeek = d >= startOfWeek && d <= endOfWeek;
      const taskIsOverdue = d < startOfWeek;
      const isCarryForward = taskIsOverdue && (t.status === 'TODO' || t.status === 'IN_PROGRESS');
      return isThisWeek || isCarryForward;
    });
  } else if (viewMode === 'backlog') {
    visibleTasks = tasks.filter(t => {
      if (!t.due_date) return false;
      return isOverdue(t.due_date) && (t.status === 'TODO' || t.status === 'IN_PROGRESS');
    });
  }

  return (
    <DragDropContext onDragEnd={onDragEnd}>
      <div className="board-container" style={{ position: 'relative' }}>
        {isFetchingTasks && (
          <div className="board-loader-overlay">
            <Spinner size={48} />
          </div>
        )}
        {COLUMNS.map(column => {
          const columnTasks = visibleTasks.filter(t => t.status === column.id);

          return (
            <div key={column.id} className="kanban-column">
              <div className="column-header">
                <span>{column.title}</span>
                <span className="column-count">{columnTasks.length}</span>
              </div>
              
              <Droppable droppableId={column.id}>
                {(provided) => (
                  <div 
                    className="task-list"
                    ref={provided.innerRef}
                    {...provided.droppableProps}
                  >
                    {columnTasks.map((task, index) => {
                      const taskOverdue = (column.id === 'TODO' || column.id === 'IN_PROGRESS') && isOverdue(task.due_date);
                      const isTaskThisWeek = (() => {
                        if (!task.due_date) return false;
                        const d = new Date(task.due_date);
                        return d >= startOfWeek && d <= endOfWeek;
                      })();
                      
                      const dateBadgeClass = taskOverdue ? 'overdue-badge' : (isTaskThisWeek ? 'current-week-badge' : '');
                      const projectName = getProjectName(task.project_id);
                      
                      return (
                      <Draggable key={task.id} draggableId={task.id.toString()} index={index}>
                        {(provided, snapshot) => (
                          <div
                            ref={provided.innerRef}
                            {...provided.draggableProps}
                            {...provided.dragHandleProps}
                            className={`task-card ${snapshot.isDragging ? 'dragging' : ''}`}
                            style={{
                              ...provided.draggableProps.style, // CRITICAL: Only use provided style to fix drag-and-drop bug
                            }}
                          >
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                              <h4 className="task-title" style={{ margin: 0, paddingRight: '10px' }}>{task.title}</h4>
                              <button className="icon-btn text-btn" style={{ padding: 0 }} onClick={() => onEditTask(task)}>
                                <Edit2 size={14} />
                              </button>
                            </div>
                            
                            <div className="task-project-badge" style={{ 
                                color: '#ffffff', 
                                backgroundColor: getProjectColor(task.project_id) 
                              }}>
                                <Folder size={10} /> {projectName}
                              </div>

                            <div className="task-meta">
                              {task.due_date && (
                                <span className={`task-date-badge ${dateBadgeClass}`}>
                                  <CalIcon size={12} /> {task.due_date}
                                </span>
                              )}
                              {task.story_points > 0 && (
                                <span className="task-points">{task.story_points} pts</span>
                              )}
                            </div>
                          </div>
                        )}
                      </Draggable>
                    )})}
                    {provided.placeholder}
                  </div>
                )}
              </Droppable>
            </div>
          );
        })}
      </div>
    </DragDropContext>
  );
};

export default Board;

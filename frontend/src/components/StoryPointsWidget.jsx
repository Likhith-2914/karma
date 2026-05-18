import React, { useState, useEffect } from 'react';
import { Target, Trophy, Edit3, Check } from 'lucide-react';

const StoryPointsWidget = ({ tasks, sprintSettings, onStartSprint }) => {
  const [target, setTarget] = useState(60);
  const [isEditing, setIsEditing] = useState(false);
  const [tempTarget, setTempTarget] = useState(60);

  useEffect(() => {
    if (sprintSettings) {
      setTarget(sprintSettings.target_points || 60);
      setTempTarget(sprintSettings.target_points || 60);
    }
  }, [sprintSettings]);

  const handleSaveTarget = () => {
    const newTarget = parseInt(tempTarget, 10) || 60;
    setTarget(newTarget);
    setIsEditing(false);
    // If they change it midway, we should probably update the backend too, 
    // but the backend only updates on 'StartSprint'. 
    // If it's Sunday, we're not technically starting a new sprint. We just want to update target.
    // For simplicity, let's allow onStartSprint to act as an update as well.
    onStartSprint(newTarget);
  };

  const today = new Date();
  const isMonday = today.getDay() === 1;
  const isSunday = today.getDay() === 0;
  
  const isSprintActive = sprintSettings && new Date(sprintSettings.sprint_end_date) > today;
  // If it's Monday and the current active sprint from the database ended yesterday (or before), 
  // it means they haven't started this week's sprint yet.
  const canStartSprint = isMonday && (!isSprintActive);

  // Calculate story points
  const points = {
    TODO: 0,
    IN_PROGRESS: 0,
    COMPLETED: 0,
    CANCELLED: 0
  };

  tasks.forEach(task => {
    if (task.status && points[task.status] !== undefined) {
      points[task.status] += (task.story_points || 0);
    }
  });

  const totalPoints = points.TODO + points.IN_PROGRESS + points.COMPLETED + points.CANCELLED;
  const isGoalReached = points.COMPLETED >= target;

  return (
    <div className={`story-points-widget ${isGoalReached ? 'goal-reached' : ''}`}>
      <div className="sp-header">
        <div className="sp-title">
          {isGoalReached ? <Trophy size={18} className="trophy-icon" /> : <Target size={18} />}
          <span>Weekly Target</span>
          {canStartSprint && (
            <button 
              className="primary-btn btn-small" 
              style={{ marginLeft: '1rem', padding: '0.2rem 0.5rem', fontSize: '0.75rem' }}
              onClick={() => onStartSprint(parseInt(tempTarget, 10) || 60)}
            >
              Start Sprint
            </button>
          )}
        </div>
        <div className="sp-target-control">
          {isEditing ? (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <input 
                type="number" 
                value={tempTarget} 
                onChange={(e) => setTempTarget(e.target.value)} 
                className="inline-input sp-input"
                autoFocus
              />
              <button onClick={handleSaveTarget} className="icon-btn text-btn" style={{ color: 'var(--success)', padding: 0 }}>
                <Check size={16} />
              </button>
            </div>
          ) : (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span className="sp-target-value">{points.COMPLETED} / {target} pts</span>
              {(isSunday || canStartSprint) && (
                <button onClick={() => setIsEditing(true)} className="icon-btn text-btn" style={{ padding: 0 }} title="Edit Target">
                  <Edit3 size={14} />
                </button>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="sp-breakdown">
        <div className="sp-item" title="To Do">
          <span className="sp-dot dot-todo"></span>
          <span className="sp-val">{points.TODO}</span>
        </div>
        <div className="sp-item" title="In Progress">
          <span className="sp-dot dot-progress"></span>
          <span className="sp-val">{points.IN_PROGRESS}</span>
        </div>
        <div className="sp-item" title="Completed">
          <span className="sp-dot dot-done"></span>
          <span className="sp-val">{points.COMPLETED}</span>
        </div>
        <div className="sp-item" title="Cancelled">
          <span className="sp-dot dot-cancelled"></span>
          <span className="sp-val">{points.CANCELLED}</span>
        </div>
        <div className="sp-item total-item" title="Total Planned">
          <span className="sp-label">Total</span>
          <span className="sp-val">{totalPoints}</span>
        </div>
      </div>
    </div>
  );
};

export default StoryPointsWidget;

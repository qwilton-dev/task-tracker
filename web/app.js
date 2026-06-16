const API = '/api/v1';
let token = localStorage.getItem('token');
let currentProject = null;
let currentWsSlug = null;
let currentWsId = null;
let workspaceLabels = [];
let workspaces = [];
let eventSource = null;
const userCache = {};

async function api(method, path, body) {
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = `Bearer ${token}`;
    const res = await fetch(API + path, { method, headers, body: body ? JSON.stringify(body) : undefined });
    if (res.status === 204) return null;
    const text = await res.text();
    if (!text) return null;
    const data = JSON.parse(text);
    if (!res.ok) throw new Error(data.error?.message || 'Request failed');
    return data;
}

function showLogin() {
    document.getElementById('login-form').classList.remove('hidden');
    document.getElementById('register-form').classList.add('hidden');
}

function showRegister() {
    document.getElementById('login-form').classList.add('hidden');
    document.getElementById('register-form').classList.remove('hidden');
}

async function login() {
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    try {
        const data = await api('POST', '/auth/login', { email, password });
        token = data.access_token;
        localStorage.setItem('token', token);
        showApp();
    } catch (e) {
        alert(e.message);
    }
}

async function register() {
    const name = document.getElementById('reg-name').value;
    const email = document.getElementById('reg-email').value;
    const password = document.getElementById('reg-password').value;
    try {
        await api('POST', '/auth/register', { name, email, password });
        alert('Registered! Now login.');
        showLogin();
    } catch (e) {
        alert(e.message);
    }
}

function logout() {
    disconnectSSE();
    token = null;
    localStorage.removeItem('token');
    document.getElementById('auth-forms').classList.remove('hidden');
    document.getElementById('main-content').classList.add('hidden');
    document.getElementById('user-info').classList.add('hidden');
}

async function showApp() {
    try {
        const user = await api('GET', '/me');
        document.getElementById('user-name').textContent = user.name || user.email;
        document.getElementById('auth-forms').classList.add('hidden');
        document.getElementById('main-content').classList.remove('hidden');
        document.getElementById('user-info').classList.remove('hidden');
        await loadWorkspaces();
    } catch {
        logout();
    }
}

async function loadWorkspaces() {
    try {
        const ws = await api('GET', '/workspaces');
        workspaces = Array.isArray(ws) ? ws : [];
        const sel = document.getElementById('workspace-select');
        sel.innerHTML = '<option value="">Select workspace</option>';
        workspaces.forEach(w => {
            sel.innerHTML += `<option value="${w.slug}" data-id="${w.id}">${w.name}</option>`;
        });
    } catch (e) {
        console.error(e);
    }
}

async function createWorkspace() {
    const name = document.getElementById('ws-name').value.trim();
    const slug = document.getElementById('ws-slug').value.trim();
    if (!name || !slug) return alert('Name and slug required');
    try {
        await api('POST', '/workspaces', { name, slug });
        closeModal();
        await loadWorkspaces();
    } catch (e) {
        alert(e.message);
    }
}

function showCreateWorkspace() {
    const modal = document.getElementById('modal');
    document.getElementById('modal-body').innerHTML = `
        <h2>New Workspace</h2>
        <input type="text" id="ws-name" placeholder="Name">
        <input type="text" id="ws-slug" placeholder="Slug (e.g. my-team)">
        <button onclick="createWorkspace()">Create</button>
    `;
    modal.classList.remove('hidden');
}

async function loadProjects() {
    const slug = document.getElementById('workspace-select').value;
    currentWsSlug = slug;
    const opt = document.getElementById('workspace-select').selectedOptions[0];
    currentWsId = opt?.dataset?.id || null;
    if (!slug) {
        document.getElementById('project-select-list').innerHTML = '<option value="">Select project</option>';
        document.getElementById('board').classList.add('hidden');
        workspaceLabels = [];
        return;
    }
    try {
        const ps = await api('GET', `/workspaces/${currentWsId}/projects`);
        const sel = document.getElementById('project-select-list');
        sel.innerHTML = '<option value="">Select project</option>';
        (Array.isArray(ps) ? ps : []).forEach(p => {
            sel.innerHTML += `<option value="${p.id}" data-key="${p.key}">${p.key} - ${p.name}</option>`;
        });
        await loadLabels();
    } catch (e) {
        console.error(e);
    }
}

async function loadLabels() {
    if (!currentWsId) { workspaceLabels = []; return; }
    try {
        const labels = await api('GET', `/workspaces/${currentWsId}/labels`);
        workspaceLabels = Array.isArray(labels) ? labels : [];
    } catch { workspaceLabels = []; }
}

function showCreateProject() {
    if (!currentWsSlug) return alert('Select a workspace first');
    const modal = document.getElementById('modal');
    document.getElementById('modal-body').innerHTML = `
        <h2>New Project</h2>
        <input type="text" id="proj-name" placeholder="Name">
        <button onclick="createProject()">Create</button>
    `;
    modal.classList.remove('hidden');
}

async function createProject() {
    const name = document.getElementById('proj-name').value.trim();
    if (!name) return alert('Name is required');
    try {
        await api('POST', `/workspaces/${currentWsId}/projects`, { name });
        closeModal();
        await loadProjects();
    } catch (e) {
        alert(e.message);
    }
}

function showManageLabels() {
    if (!currentWsSlug) return alert('Select a workspace first');
    const modal = document.getElementById('modal');
    const listHtml = workspaceLabels.length
        ? workspaceLabels.map(l => `<div class="label-item" style="display:flex;align-items:center;gap:8px;padding:6px 0"><span class="label-pill" style="background:${l.color};color:#fff;padding:2px 8px;border-radius:12px;font-size:12px">${escapeHtml(l.name)}</span></div>`).join('')
        : '<p style="color:#8b949e;font-size:14px">No labels yet</p>';

    document.getElementById('modal-body').innerHTML = `
        <h2>Manage Labels</h2>
        <div id="labels-list">${listHtml}</div>
        <div style="margin-top:16px;border-top:1px solid #30363d;padding-top:12px">
            <input type="text" id="new-label-name" placeholder="Label name">
            <div style="display:flex;gap:8px;align-items:center;margin-bottom:8px">
                <label style="font-size:12px;color:#8b949e">Color</label>
                <input type="color" id="new-label-color" value="#58a6ff" style="width:40px;height:32px;padding:2px;border:1px solid #30363d;background:#0d1117;border-radius:4px;cursor:pointer">
            </div>
            <button onclick="createLabel()">Add Label</button>
        </div>
    `;
    modal.classList.remove('hidden');
}

async function createLabel() {
    const name = document.getElementById('new-label-name').value.trim();
    const color = document.getElementById('new-label-color').value;
    if (!name) return alert('Name is required');
    try {
        await api('POST', `/workspaces/${currentWsId}/labels`, { name, color });
        await loadLabels();
        showManageLabels();
    } catch (e) {
        alert(e.message);
    }
}

async function attachLabel(issueId) {
    const select = document.getElementById('attach-label-select');
    const labelId = select.value;
    if (!labelId) return;
    try {
        await api('POST', `/issues/${issueId}/labels/${labelId}`);
        const issue = await api('GET', `/issues/${issueId}`);
        showIssueDetail(issue);
    } catch (e) {
        alert(e.message);
    }
}

async function detachLabel(issueId, labelId) {
    try {
        await api('DELETE', `/issues/${issueId}/labels/${labelId}`);
        const issue = await api('GET', `/issues/${issueId}`);
        showIssueDetail(issue);
    } catch (e) {
        alert(e.message);
    }
}

async function showManageMembers() {
    if (!currentWsId) return alert('Select a workspace first');
    const modal = document.getElementById('modal');
    let membersHtml = '<p style="color:#8b949e;font-size:14px">Loading...</p>';
    document.getElementById('modal-body').innerHTML = `
        <h2>Workspace Members</h2>
        <div id="members-list">${membersHtml}</div>
        <div style="margin-top:16px;border-top:1px solid #30363d;padding-top:12px">
            <input type="text" id="new-member-id" placeholder="User ID">
            <select id="new-member-role">
                <option value="viewer">Viewer</option>
                <option value="member">Member</option>
                <option value="owner">Owner</option>
            </select>
            <button onclick="addMember()">Add Member</button>
        </div>
        <div style="margin-top:16px;border-top:1px solid #30363d;padding-top:12px">
            <h3 style="font-size:14px;color:#8b949e;margin-bottom:8px">Invite Member</h3>
            <input type="email" id="invite-email" placeholder="Email">
            <select id="invite-role">
                <option value="member">Member</option>
                <option value="viewer">Viewer</option>
            </select>
            <button onclick="sendInvite()">Send Invite</button>
        </div>
    `;
    modal.classList.remove('hidden');
    await loadMembers();
}

async function loadMembers() {
    try {
        const members = await api('GET', `/workspaces/${currentWsId}/members`);
        const list = document.getElementById('members-list');
        if (!list) return;
        const arr = Array.isArray(members) ? members : [];
        if (!arr.length) {
            list.innerHTML = '<p style="color:#8b949e;font-size:14px">No members</p>';
            return;
        }
        list.innerHTML = arr.map(m => `
            <div class="member-item" style="display:flex;align-items:center;justify-content:space-between;padding:8px 0;border-bottom:1px solid #21262d">
                <div>
                    <span style="color:#f0f6fc;font-size:14px">${escapeHtml(m.user_name || m.user_id)}</span>
                    <span class="label-pill" style="background:#21262d;color:#8b949e;padding:2px 8px;border-radius:12px;font-size:11px;margin-left:8px">${m.role}</span>
                </div>
                <div style="display:flex;gap:4px">
                    <select class="role-select" data-user="${m.user_id}" style="width:auto;padding:2px 6px;font-size:12px;margin-bottom:0" onchange="updateMemberRole('${m.user_id}', this.value)">
                        ${['viewer','member','owner'].map(r => `<option value="${r}" ${r === m.role ? 'selected' : ''}>${r}</option>`).join('')}
                    </select>
                    <button class="secondary" style="padding:2px 8px;font-size:11px" onclick="removeMember('${m.user_id}')">Remove</button>
                </div>
            </div>
        `).join('');
    } catch (e) {
        const list = document.getElementById('members-list');
        if (list) list.innerHTML = `<p style="color:#f85149;font-size:14px">${e.message}</p>`;
    }
}

async function addMember() {
    const userId = document.getElementById('new-member-id').value.trim();
    const role = document.getElementById('new-member-role').value;
    if (!userId) return alert('User ID is required');
    try {
        await api('POST', `/workspaces/${currentWsId}/members`, { user_id: userId, role });
        document.getElementById('new-member-id').value = '';
        await loadMembers();
    } catch (e) {
        alert(e.message);
    }
}

async function updateMemberRole(userId, role) {
    try {
        await api('PATCH', `/workspaces/${currentWsId}/members/${userId}`, { role });
    } catch (e) {
        alert(e.message);
        await loadMembers();
    }
}

async function removeMember(userId) {
    if (!confirm('Remove this member?')) return;
    try {
        await api('DELETE', `/workspaces/${currentWsId}/members/${userId}`);
        await loadMembers();
    } catch (e) {
        alert(e.message);
    }
}

async function sendInvite() {
    const email = document.getElementById('invite-email').value.trim();
    const role = document.getElementById('invite-role').value;
    if (!email) return alert('Email is required');
    try {
        await api('POST', `/workspaces/${currentWsId}/invites`, { email, role });
        alert('Invite sent!');
        document.getElementById('invite-email').value = '';
    } catch (e) {
        alert(e.message);
    }
}

async function loadIssues() {
    const projectId = document.getElementById('project-select-list').value;
    if (!projectId) {
        document.getElementById('board').classList.add('hidden');
        disconnectSSE();
        return;
    }
    currentProject = {
        id: projectId,
        key: document.getElementById('project-select-list').selectedOptions[0]?.dataset.key || ''
    };
    try {
        const issues = await api('GET', `/projects/${projectId}/issues`);
        renderBoard(Array.isArray(issues) ? issues : []);
        document.getElementById('board').classList.remove('hidden');
        connectSSE(projectId);
    } catch (e) {
        console.error(e);
    }
}

function connectSSE(projectId) {
    disconnectSSE();
    eventSource = new EventSource(`${API}/projects/${projectId}/events?token=${token}`);

    eventSource.addEventListener('issue.created', (e) => {
        const issue = JSON.parse(e.data);
        addIssueToBoard(issue);
    });

    eventSource.addEventListener('issue.updated', (e) => {
        const issue = JSON.parse(e.data);
        updateIssueOnBoard(issue);
    });

    eventSource.addEventListener('issue.moved', (e) => {
        const issue = JSON.parse(e.data);
        moveIssueOnBoard(issue);
    });

    eventSource.addEventListener('issue.deleted', (e) => {
        const issue = JSON.parse(e.data);
        removeIssueFromBoard(issue.id);
    });

    eventSource.addEventListener('comment.added', async (e) => {
        const comment = JSON.parse(e.data);
        if (document.getElementById('modal').classList.contains('hidden')) return;
        const issueId = document.getElementById('new-comment')?.closest('[data-issue-id]')?.dataset.issueId;
        if (issueId === comment.issue_id) {
            const authorName = await resolveUserName(comment.author_id);
            addCommentToModal({ ...comment, author_name: authorName });
        }
    });

    eventSource.onerror = () => {
        console.warn('SSE connection lost, reconnecting...');
    };
}

function disconnectSSE() {
    if (eventSource) {
        eventSource.close();
        eventSource = null;
    }
}

function addIssueToBoard(issue) {
    const col = document.querySelector(`.column[data-status="${issue.status}"] .issues`);
    if (!col) return;
    const card = document.createElement('div');
    card.className = 'issue-card';
    card.draggable = true;
    card.dataset.id = issue.id;
    card.innerHTML = `
        <div class="key">${currentProject.key}-${issue.number}</div>
        <div class="title">${escapeHtml(issue.title)}</div>
        <div class="meta">
            <span class="priority priority-${issue.priority}">${issue.priority}</span>
        </div>
    `;
    card.addEventListener('dragstart', onDragStart);
    card.addEventListener('click', () => showIssueDetail(issue));
    col.appendChild(card);
}

function updateIssueOnBoard(issue) {
    const card = document.querySelector(`.issue-card[data-id="${issue.id}"]`);
    if (!card) return addIssueToBoard(issue);
    card.querySelector('.title').textContent = issue.title;
    const pri = card.querySelector('.priority');
    pri.textContent = issue.priority;
    pri.className = `priority priority-${issue.priority}`;
}

function moveIssueOnBoard(issue) {
    const card = document.querySelector(`.issue-card[data-id="${issue.id}"]`);
    if (!card) return addIssueToBoard(issue);
    const col = document.querySelector(`.column[data-status="${issue.status}"] .issues`);
    if (col) col.appendChild(card);
}

function removeIssueFromBoard(issueId) {
    const card = document.querySelector(`.issue-card[data-id="${issueId}"]`);
    if (card) card.remove();
}

function addCommentToModal(comment) {
    const list = document.getElementById('comments-list');
    if (!list) return;
    const noComments = list.querySelector('p');
    if (noComments) noComments.remove();
    const authorName = comment.author_name || comment.author_id;
    const div = document.createElement('div');
    div.className = 'comment';
    div.innerHTML = `
        <div class="author">${escapeHtml(authorName)}</div>
        <div class="body">${escapeHtml(comment.body)}</div>
        <div class="time">${new Date(comment.created_at).toLocaleString()}</div>
    `;
    list.appendChild(div);
}

function renderBoard(issues) {
    const columns = { backlog: [], todo: [], in_progress: [], review: [], done: [] };
    issues.forEach(i => {
        if (columns[i.status]) columns[i.status].push(i);
    });

    document.querySelectorAll('.column').forEach(col => {
        const status = col.dataset.status;
        const container = col.querySelector('.issues');
        container.innerHTML = '';
        columns[status].forEach(issue => {
            const card = document.createElement('div');
            card.className = 'issue-card';
            card.draggable = true;
            card.dataset.id = issue.id;
            card.innerHTML = `
                <div class="key">${currentProject.key}-${issue.number}</div>
                <div class="title">${escapeHtml(issue.title)}</div>
                <div class="meta">
                    <span class="priority priority-${issue.priority}">${issue.priority}</span>
                </div>
            `;
            card.addEventListener('dragstart', onDragStart);
            card.addEventListener('click', () => showIssueDetail(issue));
            container.appendChild(card);
        });

        col.ondragover = (e) => { e.preventDefault(); e.dataTransfer.dropEffect = 'move'; };
        col.ondrop = async (e) => {
            e.preventDefault();
            const newStatus = col.dataset.status;
            if (!draggedId || !newStatus) return;
            try {
                await api('PATCH', `/issues/${draggedId}/move`, { status: newStatus, position: 0 });
                await loadIssues();
            } catch (err) { alert(err.message); }
            draggedId = null;
        };
    });
}

let draggedId = null;

function onDragStart(e) {
    draggedId = e.target.dataset.id;
    e.target.classList.add('dragging');
    e.dataTransfer.effectAllowed = 'move';
}

function showCreateIssue() {
    if (!currentProject) return alert('Select a project first');
    const modal = document.getElementById('modal');
    document.getElementById('modal-body').innerHTML = `
        <h2>New Issue</h2>
        <input type="text" id="new-issue-title" placeholder="Title">
        <textarea id="new-issue-desc" placeholder="Description (optional)"></textarea>
        <button onclick="createIssue()">Create</button>
    `;
    modal.classList.remove('hidden');
}

async function createIssue() {
    const title = document.getElementById('new-issue-title').value;
    const description = document.getElementById('new-issue-desc').value;
    if (!title) return alert('Title is required');
    try {
        await api('POST', `/projects/${currentProject.id}/issues`, { title, description });
        closeModal();
        await loadIssues();
    } catch (e) {
        alert(e.message);
    }
}

async function showIssueDetail(issue) {
    const modal = document.getElementById('modal');
    let commentsHtml = '';
    try {
        const comments = await api('GET', `/issues/${issue.id}/comments`);
        const arr = Array.isArray(comments) ? comments : [];
        for (const c of arr) {
            const authorName = await resolveUserName(c.author_id);
            commentsHtml += `
                <div class="comment">
                    <div class="author">${escapeHtml(authorName)}</div>
                    <div class="body">${escapeHtml(c.body)}</div>
                    <div class="time">${new Date(c.created_at).toLocaleString()}</div>
                </div>
            `;
        }
    } catch {}

    let issueLabels = [];
    try {
        const labels = await api('GET', `/issues/${issue.id}/labels`);
        issueLabels = Array.isArray(labels) ? labels : [];
    } catch {}

    let activityHtml = '';
    try {
        const activity = await api('GET', `/issues/${issue.id}/activity`);
        const arr = Array.isArray(activity) ? activity : [];
        if (arr.length) {
            const entries = await Promise.all(arr.map(async e => {
                const actorName = await resolveUserName(e.actor_id);
                const desc = formatActivityEvent(e);
                return `<div style="display:flex;gap:8px;padding:6px 0;font-size:12px;color:#8b949e;border-bottom:1px solid #21262d">
                    <span style="color:#484f58;white-space:nowrap">${new Date(e.created_at).toLocaleString()}</span>
                    <span style="color:#58a6ff">${escapeHtml(actorName)}</span>
                    <span style="color:#c9d1d9">${desc}</span>
                </div>`;
            }));
            activityHtml = entries.join('');
        }
    } catch {}

    const issueLabelIds = new Set(issueLabels.map(l => l.id));
    const availableLabels = workspaceLabels.filter(l => !issueLabelIds.has(l.id));

    const attachedLabelsHtml = issueLabels.length
        ? issueLabels.map(l => `<span class="label-pill" style="background:${l.color};color:#fff;padding:2px 8px;border-radius:12px;font-size:11px;display:inline-flex;align-items:center;gap:4px">${escapeHtml(l.name)}<span style="cursor:pointer;opacity:0.7" onclick="detachLabel('${issue.id}','${l.id}')">&times;</span></span>`).join(' ')
        : '<span style="color:#484f58;font-size:12px">No labels</span>';

    const attachOptionsHtml = availableLabels.length
        ? availableLabels.map(l => `<option value="${l.id}">${escapeHtml(l.name)}</option>`).join('')
        : '<option disabled>No more labels</option>';

    document.getElementById('modal-body').innerHTML = `
        <h2>${currentProject.key}-${issue.number}</h2>
        <div class="issue-detail">
            <div class="field">
                <label>Title</label>
                <input type="text" id="edit-title" value="${escapeHtml(issue.title)}">
            </div>
            <div class="field">
                <label>Description</label>
                <textarea id="edit-desc">${escapeHtml(issue.description || '')}</textarea>
            </div>
            <div class="field">
                <label>Priority</label>
                <select id="edit-priority">
                    ${['none','low','medium','high','urgent'].map(p =>
                        `<option value="${p}" ${p === issue.priority ? 'selected' : ''}>${p}</option>`
                    ).join('')}
                </select>
            </div>
            <div class="field">
                <label>Status</label>
                <select id="edit-status">
                    ${['backlog','todo','in_progress','review','done'].map(s =>
                        `<option value="${s}" ${s === issue.status ? 'selected' : ''}>${s}</option>`
                    ).join('')}
                </select>
            </div>
            <div class="field">
                <label>Labels</label>
                <div id="issue-labels" style="display:flex;flex-wrap:wrap;gap:4px;margin-bottom:8px">${attachedLabelsHtml}</div>
                <div style="display:flex;gap:6px">
                    <select id="attach-label-select" style="margin-bottom:0;flex:1">${attachOptionsHtml}</select>
                    <button class="secondary" style="margin-bottom:0;padding:6px 12px" onclick="attachLabel('${issue.id}')">Attach</button>
                </div>
            </div>
            <button onclick="updateIssue('${issue.id}')">Save</button>
            <button class="secondary" onclick="deleteIssue('${issue.id}')" style="margin-left:8px">Delete</button>
        </div>
        <div class="comments-section">
            <h3>Comments</h3>
            <div id="comments-list">${commentsHtml || '<p style="color:#8b949e;font-size:14px">No comments</p>'}</div>
            <div style="margin-top:12px">
                <textarea id="new-comment" placeholder="Add comment..."></textarea>
                <button onclick="addComment('${issue.id}')" style="margin-top:8px">Comment</button>
            </div>
        </div>
        <div class="comments-section" style="margin-top:16px">
            <h3>Activity</h3>
            <div id="activity-list">${activityHtml || '<p style="color:#8b949e;font-size:14px">No activity</p>'}</div>
        </div>
    `;
    modal.classList.remove('hidden');
}

async function updateIssue(id) {
    const title = document.getElementById('edit-title').value;
    const description = document.getElementById('edit-desc').value;
    const priority = document.getElementById('edit-priority').value;
    const status = document.getElementById('edit-status').value;
    try {
        await api('PATCH', `/issues/${id}`, { title, description, priority });
        if (status) await api('PATCH', `/issues/${id}/move`, { status, position: 0 });
        closeModal();
        await loadIssues();
    } catch (e) {
        alert(e.message);
    }
}

async function deleteIssue(id) {
    if (!confirm('Delete this issue?')) return;
    try {
        await api('DELETE', `/issues/${id}`);
        closeModal();
        await loadIssues();
    } catch (e) {
        alert(e.message);
    }
}

async function addComment(issueId) {
    const body = document.getElementById('new-comment').value;
    if (!body) return;
    try {
        await api('POST', `/issues/${issueId}/comments`, { body });
        const issue = await api('GET', `/issues/${issueId}`);
        showIssueDetail(issue);
    } catch (e) {
        alert(e.message);
    }
}

function closeModal() {
    document.getElementById('modal').classList.add('hidden');
}

function escapeHtml(s) {
    if (!s) return '';
    const d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
}

async function resolveUserName(userId) {
    if (!userId) return 'system';
    if (userCache[userId]) return userCache[userId];
    try {
        const members = await api('GET', `/workspaces/${currentWsId}/members`);
        const arr = Array.isArray(members) ? members : [];
        arr.forEach(m => { userCache[m.user_id] = m.user_name || m.user_id; });
    } catch {}
    return userCache[userId] || userId.slice(0, 8);
}

function formatActivityEvent(e) {
    const eventType = e.type || e.event_type || 'unknown';
    const p = typeof e.payload === 'string' ? JSON.parse(e.payload) : (e.payload || {});
    switch (eventType) {
        case 'issue.created': return `created issue "${p.title || ''}"`;
        case 'issue.updated': return `updated issue "${p.title || ''}"`;
        case 'issue.moved': return `moved issue "${p.title || ''}"`;
        case 'issue.deleted': return `deleted issue "${p.title || ''}"`;
        case 'comment.added': return `added a comment`;
        case 'issue.label_added': return `added label "${p.name || ''}"`;
        case 'issue.label_removed': return `removed label "${p.name || ''}"`;
        default: return eventType;
    }
}

document.getElementById('modal').addEventListener('click', (e) => {
    if (e.target === e.currentTarget) closeModal();
});

if (token) showApp();

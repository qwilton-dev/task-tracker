const API = '/api/v1';
let token = localStorage.getItem('token');
let currentProject = null;
let currentWsSlug = null;

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
        const sel = document.getElementById('workspace-select');
        sel.innerHTML = '<option value="">Select workspace</option>';
        (Array.isArray(ws) ? ws : []).forEach(w => {
            sel.innerHTML += `<option value="${w.slug}">${w.name}</option>`;
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
    if (!slug) {
        document.getElementById('project-select-list').innerHTML = '<option value="">Select project</option>';
        document.getElementById('board').classList.add('hidden');
        return;
    }
    try {
        const ps = await api('GET', `/workspaces/${slug}/projects`);
        const sel = document.getElementById('project-select-list');
        sel.innerHTML = '<option value="">Select project</option>';
        (Array.isArray(ps) ? ps : []).forEach(p => {
            sel.innerHTML += `<option value="${p.id}" data-key="${p.key}">${p.key} - ${p.name}</option>`;
        });
    } catch (e) {
        console.error(e);
    }
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
        await api('POST', `/workspaces/${currentWsSlug}/projects`, { name });
        closeModal();
        await loadProjects();
    } catch (e) {
        alert(e.message);
    }
}

async function loadIssues() {
    const projectId = document.getElementById('project-select-list').value;
    if (!projectId) {
        document.getElementById('board').classList.add('hidden');
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
    } catch (e) {
        console.error(e);
    }
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
        (Array.isArray(comments) ? comments : []).forEach(c => {
            commentsHtml += `
                <div class="comment">
                    <div class="author">${c.author_id}</div>
                    <div class="body">${escapeHtml(c.body)}</div>
                    <div class="time">${new Date(c.created_at).toLocaleString()}</div>
                </div>
            `;
        });
    } catch {}

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

document.getElementById('modal').addEventListener('click', (e) => {
    if (e.target === e.currentTarget) closeModal();
});

if (token) showApp();

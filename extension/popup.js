const API_BASE_URL = 'http://localhost:8080'

const urlEl = document.getElementById('url')
const titleEl = document.getElementById('title')
const memoEl = document.getElementById('memo')
const tagsEl = document.getElementById('tags')
const formEl = document.getElementById('save-form')
const buttonEl = document.getElementById('save-button')
const messageEl = document.getElementById('message')

let currentTab = null

async function init() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true })
  currentTab = tab
  urlEl.textContent = tab?.url ?? ''
  titleEl.value = tab?.title ?? ''

  if (!tab?.url) {
    showMessage('このページは保存できません', true)
    buttonEl.disabled = true
  }
}

function showMessage(text, isError) {
  messageEl.textContent = text
  messageEl.hidden = false
  messageEl.classList.toggle('error', isError)
}

formEl.addEventListener('submit', async (event) => {
  event.preventDefault()
  if (!currentTab?.url) return

  buttonEl.disabled = true
  showMessage('保存中...', false)

  const tags = tagsEl.value
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag !== '')

  try {
    const res = await fetch(`${API_BASE_URL}/api/links`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        url: currentTab.url,
        title: titleEl.value.trim(),
        memo: memoEl.value.trim(),
        tags,
      }),
    })
    if (!res.ok) {
      throw new Error(`failed to save: ${res.status}`)
    }

    showMessage('保存しました', false)
    setTimeout(() => window.close(), 800)
  } catch (err) {
    console.error(err)
    showMessage('保存に失敗しました。LinkVaultのサーバーが起動しているか確認してください。', true)
    buttonEl.disabled = false
  }
})

init()

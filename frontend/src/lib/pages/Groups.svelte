<script lang="ts">
  import { onMount } from 'svelte';
  import GroupService from '$lib/services/groupService';
  import type { Group, GroupPayload } from '$lib/types';

  let groups: Group[] = [];
  let isLoading = true;
  let error: string | null = null;

  let searchTerm: string = '';
  // let sortBy: keyof Group = 'name'; // Если понадобится сортировка для групп

  // Состояние для модального окна/формы создания/редактирования
  let showModal = false;
  let currentGroup: Partial<GroupPayload> = {}; 
  let editingGroupId: string | null = null; 

  onMount(async () => {
    await loadGroups();
  });

  async function loadGroups() {
    isLoading = true;
    error = null;
    try {
      const fetchedGroups = await GroupService.getAllGroups();
      groups = fetchedGroups || [];
    } catch (err: any) {
      console.error("Failed to load groups:", err);
      error = err.message || "Не удалось загрузить группы.";
      groups = [];
    }
    isLoading = false;
  }

  $: filteredGroups = groups.filter(group => 
    (group.name?.toLowerCase() || '').includes(searchTerm.toLowerCase()) ||
    (group.description?.toLowerCase() || '').includes(searchTerm.toLowerCase())
  )/*.sort((a, b) => { // Если понадобится сортировка
    if (a[sortBy] < b[sortBy]) return -1;
    if (a[sortBy] > b[sortBy]) return 1;
    return 0;
  })*/;

  function openCreateModal() {
    editingGroupId = null;
    currentGroup = { name: '', description: '' }; 
    showModal = true;
  }

  function openEditModal(group: Group) {
    editingGroupId = group.id;
    currentGroup = { ...group }; 
    showModal = true;
  }

  async function handleDelete(groupId: string) {
    if (!confirm('Вы уверены, что хотите удалить эту группу? Все связанные контакты останутся, но будут откреплены от этой группы.')) return;
    isLoading = true; 
    try {
      await GroupService.deleteGroup(groupId);
      await loadGroups(); 
    } catch (err: any) {
      console.error("Failed to delete group:", err);
      error = err.message || "Не удалось удалить группу.";
    }
    isLoading = false;
  }
  
  async function handleFormSubmit() {
    if (!currentGroup.name) {
        alert("Название группы обязательно для заполнения.");
        return;
    }
    isLoading = true;
    error = null;
    try {
      if (editingGroupId) {
        await GroupService.updateGroup(editingGroupId, currentGroup as GroupPayload);
      } else {
        await GroupService.createGroup(currentGroup as GroupPayload);
      }
      showModal = false;
      await loadGroups(); 
    } catch (err: any) {
      console.error("Failed to save group:", err);
      error = err.message || "Не удалось сохранить группу.";
    }
    isLoading = false;
  }

</script>

<div class="groups-page">
  <h2>Управление группами</h2>

  <div class="controls">
    <input type="text" placeholder="Поиск по названию или описанию..." bind:value={searchTerm} disabled={isLoading}>
    <button on:click={openCreateModal} disabled={isLoading}>Добавить группу</button>
  </div>

  {#if isLoading && groups.length === 0}
    <p>Загрузка групп...</p>
  {:else if error}
    <p class="error-message">Ошибка: {error}</p>
    <button on:click={loadGroups}>Попробовать снова</button>
  {:else if filteredGroups.length > 0}
    <ul class="group-list">
      {#each filteredGroups as group (group.id)}
        <li class="group-item">
          <h3>{group.name}</h3>
          {#if group.description}<p>{group.description}</p>{/if}
          <!-- TODO: Отображение участников группы, если это необходимо -->
          <div class="actions">
            <button class="edit" on:click={() => openEditModal(group)}>Редактировать</button>
            <button class="delete" on:click={() => handleDelete(group.id)}>Удалить</button>
          </div>
        </li>
      {/each}
    </ul>
  {:else}
     <p>Группы не найдены. <button on:click={openCreateModal}>Добавить первую группу?</button></p>
  {/if}
</div>

{#if showModal}
  <div class="modal-backdrop" on:click={() => showModal = false}></div>
  <div class="modal">
    <h3>{editingGroupId ? 'Редактировать группу' : 'Создать группу'}</h3>
    <form on:submit|preventDefault={handleFormSubmit}>
      <div class="form-group">
        <label for="groupName">Название группы*:</label>
        <input type="text" id="groupName" bind:value={currentGroup.name} required disabled={isLoading}>
      </div>
      <div class="form-group">
        <label for="groupDescription">Описание:</label>
        <textarea id="groupDescription" bind:value={currentGroup.description} disabled={isLoading}></textarea>
      </div>
      {#if error && showModal}
        <p class="error-message">{error}</p>
      {/if}
      <div class="modal-actions">
        <button type="submit" class="primary" disabled={isLoading}>{#if isLoading}Сохранение...{:else}Сохранить{/if}</button>
        <button type="button" on:click={() => showModal = false} disabled={isLoading}>Отмена</button>
      </div>
    </form>
  </div>
{/if}

<style>
  /* Стили аналогичны Contacts.svelte, но можно адаптировать */
  .groups-page {
    padding: 20px;
  }
  .controls {
    display: flex;
    gap: 10px;
    margin-bottom: 20px;
    align-items: center;
  }
  .controls input[type="text"] {
    flex-grow: 1;
    margin-bottom: 0;
  }
  .controls button {
    background-color: #1890ff;
    color: white;
    border: none;
  }
  .controls button:hover {
    background-color: #40a9ff;
  }
  .group-list {
    list-style: none;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 20px;
  }
  .group-item {
    background-color: #fff;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    display: flex;
    flex-direction: column;
  }
  .group-item h3 {
    margin-top: 0;
    margin-bottom: 8px;
    color: #007bff; 
  }
  .group-item p {
    font-size: 0.9rem;
    color: #555;
    flex-grow: 1; /* Чтобы описание занимало доступное место */
  }
  .actions {
    margin-top: 15px; /* Отступ для кнопок */
    padding-top: 10px;
    border-top: 1px solid #f0f0f0; /* Разделитель */
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }
  .actions button {
    padding: 6px 12px;
    font-size: 0.85rem;
  }
  .actions button.edit {
    background-color: #1890ff;
    color: white;
  }
  .actions button.edit:hover {
    background-color: #40a9ff;
  }
  .actions button.delete {
    background-color: #ff4d4f;
    color: white;
  }
  .actions button.delete:hover {
    background-color: #ff7875;
  }
  .modal-backdrop, .modal, .modal-actions, .error-message { /* Используем стили из Contacts.svelte, если они подходят */
    /* ... скопируйте или импортируйте общие стили модального окна ... */
  }
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0,0,0,0.5);
    z-index: 10;
  }
  .modal {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    background-color: white;
    padding: 25px;
    border-radius: 8px;
    box-shadow: 0 5px 15px rgba(0,0,0,0.3);
    z-index: 11;
    width: 90%;
    max-width: 500px;
    max-height: 90vh;
    overflow-y: auto;
  }
  .modal h3 {
    margin-top: 0;
    margin-bottom: 20px;
  }
  .modal .form-group {
    margin-bottom: 15px;
  }
  .modal-actions {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }
  .modal-actions button.primary {
    background-color: #1890ff;
    color: white;
  }
  .modal-actions button.primary:hover {
    background-color: #40a9ff;
  }
  .modal-actions button.primary:disabled,
  .modal-actions button:disabled {
    background-color: #ccc;
    cursor: not-allowed;
  }
  .error-message {
    color: red;
    margin-bottom: 10px;
  }
</style> 
# Requirements: Мультимодальная система оценки психоэмоционального состояния специалистов

**Defined:** 2026-03-24
**Core Value:** Система должна давать оператору надёжный, интерпретируемый и воспроизводимый результат обследования специалиста, основанный на полном мультимодальном анализе речевых ответов, а не на ручной субъективной оценке.

## v1 Requirements

### UI Contours

- [x] **CNTR-01**: Пользователь с ролью `operator` работает в отдельном operator-контуре с собственной навигацией, layout и без admin-элементов управления
- [x] **CNTR-02**: Пользователь с ролью `admin` работает в отдельном admin-контуре с собственной навигацией, layout и без operator-сценариев проведения обследования

### Administration: Users and Roles

- [x] **ADMN-01**: Администратор может просматривать список пользователей с их ролями и статусом доступа
- [x] **ADMN-02**: Администратор может создавать, редактировать и изменять роль пользователя через UI с серверной валидацией прав

### Questionnaire Builder

- [x] **QSTR-01**: Администратор может создавать опросник с названием, описанием и списком вопросов
- [x] **QSTR-02**: Администратор может редактировать состав, порядок и состояние публикации опросника

### System Settings

- [x] **STNG-01**: Администратор может изменять через UI политики TTL хранения аудио и параметры retry/processing, разрешённые текущей архитектурой
- [x] **STNG-02**: Экран системных настроек показывает текущие значения, валидирует ввод и даёт явный success/error feedback после сохранения

### Monitoring and Audit

- [x] **MONR-01**: Администратор может просматривать в UI базовый статус системы по health/readiness и ключевым техническим индикаторам без прямого доступа к инфраструктуре
- [x] **AUDT-01**: Администратор может просматривать и фильтровать audit log по типу события, периоду и связанным доменным объектам

### Operator Experience

- [x] **OPRX-01**: Оператор видит results screen с понятной интерпретацией результата, baseline-отклонениями, вкладом каналов и отдельно вынесенными техническими деталями
- [x] **OPRX-02**: Сценарий обследования даёт оператору явный feedback для записи, загрузки, сохранения, завершения и recoverable ошибок
- [x] **OPRX-03**: Ключевые operator и admin страницы корректно обрабатывают `loading`, `empty` и `error` состояния без blank или placeholder UI
- [x] **OPRX-04**: Ключевые edit и destructive действия в operator/admin интерфейсах требуют подтверждения там, где это необходимо, и показывают явный success/error feedback после выполнения

### Design System and Visual Completion

- [x] **DSGN-01**: Operator и admin интерфейсы используют единый набор design tokens и переиспользуемых компонентов для типографики, цветов, отступов, форм и состояний
- [x] **DSGN-02**: В shipped frontend не остаются временные заглушки, черновые элементы и визуально незавершённые экраны в основном пользовательском потоке

## v2 Requirements

### Advanced Results

- **RICH-01**: Оператор видит расширенную визуализацию результата через графики, тренды и более детализированное baseline-сравнение

### Advanced Monitoring

- **AMON-01**: Администратор может просматривать расширенную техническую статистику и более глубокие runtime-метрики в админке

## Out of Scope

| Feature | Reason |
|---------|--------|
| Изменение ML-логики каналов `text` / `acoustic` / `paralinguistic` | Этот milestone сфокусирован только на UI/admin completion |
| Изменение decision layer и real WiMi model output | Это отдельный следующий product/integration milestone |
| Добавление новых сервисов или пересборка архитектуры | Ограничение milestone: не расширять систему инфраструктурно |
| Крупный backend refactor вне нужд UI/admin | Backend меняется только точечно под контракты и пользовательский слой |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| CNTR-01 | Phase 8 | Complete |
| CNTR-02 | Phase 8 | Complete |
| ADMN-01 | Phase 9 | Complete |
| ADMN-02 | Phase 9 | Complete |
| QSTR-01 | Phase 9 | Complete |
| QSTR-02 | Phase 9 | Complete |
| STNG-01 | Phase 10 | Complete |
| STNG-02 | Phase 10 | Complete |
| MONR-01 | Phase 10 | Complete |
| AUDT-01 | Phase 10 | Complete |
| OPRX-01 | Phase 11 | Complete |
| OPRX-02 | Phase 11 | Complete |
| OPRX-03 | Phase 11 | Complete |
| OPRX-04 | Phase 11 | Complete |
| DSGN-01 | Phase 8 | Complete |
| DSGN-02 | Phase 11 | Complete |

**Coverage:**
- v1 requirements: 16 total
- Mapped to phases: 16
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-24*
*Last updated: 2026-03-25 after Phase 11 completion*

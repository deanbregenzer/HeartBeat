# Claude Git Workflow (Kurzversion)

## Branches
- `main`: Prod, **nie direkt pushen**
- `develop`: Dev-Branch
- Feature: `feature/<name>`
- Bugfix: `bugfix/<name>`
- Refactor: `refactor/<scope>`
- Docs: `docs/<topic>`

## Workflow
```bash
git checkout develop && git pull
git checkout -b feature/<name>
# Commits mit Conventional Commits
git rebase develop  # vor Push

```
Use Conventional Commits !
**Types:**  
feat | fix | docs | style | refactor | test | chore | perf | ci  

**Regeln:**  
- Subject ≤ 50 Zeichen  
- Imperativform (z. B. *add*, nicht *added*)  
- Kein Punkt am Ende  
- Optionaler Body erklärt *was/warum* 
- Footer für `Closes #123` oder `BREAKING CHANGE: ...`

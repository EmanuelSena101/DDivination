

## 🔮 **DDivination: The RPG Master's Oracle**

<p align="center">
<img src="https://img.shields.io/badge/Python-3.9+-blue.svg" alt="Python 3.9+">
<img src="https://img.shields.io/badge/License-MIT-green.svg" alt="MIT License">
<img src="https://img.shields.io/badge/Status-In%20Development-orange.svg" alt="In Development">
</p>

### ✨ **Overview & Philosophy**

**DDivination** is an **intelligent and powerful internal tool** designed to revolutionize how RPG Game Masters plan and execute their sessions. By combining advancements in **Data Science, Procedural Generation, Natural Language Processing (NLP), and Visualization**, this project aims to **mitigate the burden of preparation**, allowing creativity and storytelling to shine at the gaming table.

**Our Goal:** To **transform intuitive commands into detailed, balanced, and immersive adventures** for internal use.

Imagine the following prompt:
*"I want a dungeon in a forgotten temple, with deadly traps and a lich boss, for 4 level 5 adventurers. The atmosphere should be swampy, with aquatic monsters, and there should be rare treasure at the end."*

**DDivination** will not only interpret this complex request but will also:

* **Structure the Dungeon:** Generate a **logical map** with rooms, corridors, and points of interest.
* **Populate the Environment:** Insert **monsters, traps, and treasures**, balanced for your group.
* **Analyze Gameplay:** Evaluate **difficulty, identify paths, and assess risk/reward distribution**.
* **Enhance Narrative:** Suggest **atmospheric descriptions and lore elements** for each room and encounter.
* **Facilitate Play:** Present all content within the application, with an **option to generate a PDF map**.

**Stop endless prepping; start playing with DDivination!**

---

### 🚀 **Key Features**

**DDivination** integrates a **robust set of features** to address the demanding needs of an RPG Game Master:

**Intelligent Dungeon Generation**

* **Natural Language Commands:** Create entire dungeons from simple textual descriptions.
* **Detailed Customization:** Define themes, sizes, room types, and bosses.
* **Graph-based Structure:** Dungeons modeled as **graphs**, allowing complex interconnections.

**In-depth Dungeon Analysis**

* **Structural Visualization:** Visual representations of dungeon layouts.
* **Difficulty Metrics:** Estimated difficulty per room and entire dungeon.
* **Path Analysis:** Identify routes, optimal paths, dead ends, and backtracking.
* **Risk and Reward Distribution:** Balanced spacing of treasures and challenges.

**Dynamic Encounter Generation**

* **Contextual Encounters:** Populate rooms based on environment, party level, and theme.
* **Automatic Balancing:** Combat encounters **fairly challenging** using CR and other metrics.
* **Treasure Suggestions:** Generate **balanced loot** based on rarity and player power.
* **Narrative Elements:** Add **atmosphere, NPC motivations, and scene details** via LLMs.

**Output & Visualization**

* **Internal Web UI:** All generated content displayed within the application.
* **PDF Map Generation:** Export **visual dungeon maps** for offline use.
* **Interactive Views:** Interactive maps, room tables, and difficulty distribution graphs.

---

### 💡 **The Idea Behind It: Engineering and Magic**

**DDivination** unites the **rigor of software engineering** with the **magic of world-building**:

* **Dungeons as Data Structures:** Represent dungeons as **graphs (networkx)** with rooms as nodes and connections as edges.
* **NLP for Intuition:** **LLMs interpret Game Master's commands** into concrete parameters.
* **Robust Procedural Generation:** Custom rules and controlled randomness for **unique, coherent content**.
* **Data Science for Balance:** Algorithms to **balance encounters, loot, and adventure progression**.

---

### ⚙️ **Detailed Tech Stack**

**DDivination** uses modern technologies across its stack:

| Layer                       | Tools / Libraries                                           |
| --------------------------- | ----------------------------------------------------------- |
| **Backend / Core Logic**    | Python, FastAPI, pandas, NumPy, networkx, random            |
| **NLP (Interpretation)**    | OpenAI API / Local LLM (Mistral via Ollama), spaCy          |
| **Procedural Generation**   | Python with custom rules for rooms, paths, encounters       |
| **Database**                | MongoDB / PostgreSQL, Vector DB (FAISS / Pinecone optional) |
| **Visualization**           | matplotlib, plotly, Graphviz, pygame (future interactive)   |
| **Frontend**                | Streamlit / React / Vue                                     |
| **Export**                  | reportlab (PDF map generation)                              |
| **Infrastructure / Deploy** | Docker, GitHub Actions, Vercel / Render                     |

---

### 📂 **Project Structure**

```
ddivination/
├── generation/
│   ├── dungeon_generator.py
│   ├── encounter_generator.py
│   ├── room_types.py
│   └── assets/
├── analysis/
│   ├── difficulty_analyzer.py
│   ├── path_finder.py
│   └── loot_analyzer.py
├── nlp/
│   ├── prompt_parser.py
│   ├── llm_integrator.py
│   └── templates/
├── ui/
│   ├── app.py
│   └── static/
├── data/
│   ├── bestiary.json
│   ├── loot_tables.json
│   └── narrative_snippets.json
├── export/
│   └── pdf_exporter.py
├── tests/
│   ├── unit/
│   └── integration/
├── utils/
│   ├── decorators.py
│   └── helpers.py
├── main.py
├── config.py
├── requirements.txt
└── README.md
```

---

### 🤝 **Contributing**

We welcome **contributions** from the internal team and collaborators! Whether it's **code, documentation, ideas, or feedback**, every bit helps DDivination grow. Please refer to **CONTRIBUTING.md** for guidelines.

---

### 📜 **License**

This project is licensed under the **MIT License** - see the LICENSE file for details.

---

### 💬 **Contact**

Have **questions, suggestions, or just want to chat** about DDivination? Feel free to **open an issue** or reach out to the project maintainers!

---

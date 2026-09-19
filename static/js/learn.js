function learnSession() {
	return {
		queue: [],
		current: null,
		feedback: null,
		answer: "",
		passesRequired: 3,
		stages: {},
		totalUnits: 0,
		mastered: 0,
		total: 0,
		attempts: 0,
		correctFirst: 0,
		seen: {},
		missed: {},
		done: false,
		empty: false,
		_advanceTimer: null,

		init(root) {
			const el = root || document;
			const n = parseInt(el.dataset && el.dataset.passes, 10);
			this.passesRequired = n >= 1 && n <= 10 ? n : 3;
			const src = el.querySelector("#learn-cards");
			let cards = [];
			try {
				cards = JSON.parse(src ? src.textContent : "[]");
			} catch (e) {
				cards = [];
			}
			cards = (cards || []).filter((c) => c && (c.front !== "" || c.back !== ""));
			this.total = cards.length;
			if (this.total === 0) {
				this.empty = true;
				return;
			}
			this._cards = cards;
			this.restart();
		},

		restart() {
			if (this._advanceTimer) clearTimeout(this._advanceTimer);
			// Each term must be answered correctly passesRequired times:
			// recognition first, then written (written only for single-card decks).
			this.totalUnits = this.passesRequired * this.total;
			this.stages = {};
			const idxs = this._cards.map((_, i) => i);
			this._shuffle(idxs);
			this.queue = idxs.map((idx) => ({ idx, stage: 0 }));
			this.mastered = 0;
			this.attempts = 0;
			this.correctFirst = 0;
			this.seen = {};
			this.missed = {};
			this.done = false;
			this.feedback = null;
			this.answer = "";
			this._next();
		},

		get doneUnits() {
			let done = 0;
			for (const k in this.stages) done += this.stages[k];
			return done;
		},

		get progress() {
			if (!this.totalUnits) return 0;
			return Math.min(100, Math.round((this.doneUnits / this.totalUnits) * 100));
		},

		get progressLabel() {
			return `${this.doneUnits} of ${this.totalUnits}`;
		},

		_norm(s) {
			return (s || "").trim().replace(/\s+/g, " ").toLowerCase();
		},

		_shuffle(a) {
			for (let i = a.length - 1; i > 0; i--) {
				const j = Math.floor(Math.random() * (i + 1));
				[a[i], a[j]] = [a[j], a[i]];
			}
			return a;
		},

		// Missed terms come back a few cards later, never immediately next.
		_requeueSoon(item) {
			const pos = 1 + Math.floor(Math.random() * (Math.min(4, this.queue.length) + 1));
			this.queue.splice(Math.min(pos, this.queue.length), 0, item);
		},

		// Graduated terms re-enter at a random spot so neither
		// question order nor question type is predictable.
		_requeueLater(item) {
			const pos = Math.floor(Math.random() * (this.queue.length + 1));
			this.queue.splice(pos, 0, item);
		},

		// A correct answer moves the term one stage up. Reaching
		// passesRequired stages means mastered and leaves the queue.
		_advance(idx, stage) {
			const ns = stage + 1;
			this.stages[idx] = Math.max(this.stages[idx] || 0, ns);
			if (ns >= this.passesRequired) {
				this.mastered++;
			} else {
				this._requeueLater({ idx, stage: ns });
			}
		},

		_next() {
			if (this._advanceTimer) {
				clearTimeout(this._advanceTimer);
				this._advanceTimer = null;
			}
			this.feedback = null;
			this.answer = "";
			if (this.queue.length === 0) {
				this.current = null;
				this.done = true;
				return;
			}
			const item = this.queue.shift();
			const card = this._cards[item.idx];
			if (item.stage === 0 && this.total > 1 && Math.random() >= 0.5) {
				this.current = {
					idx: item.idx,
					stage: item.stage,
					front: card.front,
					back: card.back,
					type: "mc",
					options: this._buildOptions(item.idx),
					picked: -1,
				};
			} else if (item.stage === 0 && this.total > 1) {
				const tf = this._buildTF(item.idx);
				this.current = {
					idx: item.idx,
					stage: item.stage,
					front: card.front,
					back: card.back,
					type: "tf",
					tfStatement: tf.statement,
					tfAnswer: tf.answer,
					picked: null,
				};
			} else {
				this.current = {
					idx: item.idx,
					stage: item.stage,
					front: card.front,
					back: card.back,
					type: "written",
					picked: -1,
				};
			}
		},

		_buildOptions(idx) {
			const correct = this._cards[idx].back;
			const pool = [];
			const seen = new Set([this._norm(correct)]);
			for (let i = 0; i < this._cards.length && pool.length < 20; i++) {
				if (i === idx) continue;
				const b = this._cards[i].back;
				const k = this._norm(b);
				if (!k || seen.has(k)) continue;
				seen.add(k);
				pool.push(b);
			}
			this._shuffle(pool);
			const opts = pool.slice(0, 3).map((text) => ({ text, correct: false }));
			opts.push({ text: correct, correct: true });
			return this._shuffle(opts);
		},

		_buildTF(idx) {
			const card = this._cards[idx];
			const norm = this._norm(card.back);
			const pool = [];
			for (let i = 0; i < this._cards.length; i++) {
				if (i === idx) continue;
				const b = this._cards[i].back;
				if (!b || this._norm(b) === norm || pool.includes(b)) continue;
				pool.push(b);
			}
			const showCorrect = pool.length === 0 || Math.random() < 0.5;
			return {
				statement: showCorrect ? card.back : pool[Math.floor(Math.random() * pool.length)],
				answer: showCorrect,
			};
		},

		_firstTry(idx) {
			return !this.seen[idx];
		},

		answerMC(i) {
			if (!this.current || this.current.type !== "mc" || this.feedback) return;
			this.current.picked = i;
			const opt = this.current.options[i];
			this.attempts++;
			const first = this._firstTry(this.current.idx);
			this.seen[this.current.idx] = true;
			if (opt && opt.correct) {
				if (first) this.correctFirst++;
				this.feedback = { correct: true, message: "Nice!" };
				this._advance(this.current.idx, this.current.stage);
				this._advanceTimer = setTimeout(() => this._next(), 650);
			} else {
				this.missed[this.current.idx] = (this.missed[this.current.idx] || 0) + 1;
				this.feedback = { correct: false, message: "Not quite." };
				this._requeueSoon({ idx: this.current.idx, stage: 0 });
				this._advanceTimer = setTimeout(() => this._next(), 2200);
			}
		},

		answerTF(value) {
			if (!this.current || this.current.type !== "tf" || this.feedback) return;
			this.current.picked = value;
			this.attempts++;
			const first = this._firstTry(this.current.idx);
			this.seen[this.current.idx] = true;
			if (value === this.current.tfAnswer) {
				if (first) this.correctFirst++;
				this.feedback = { correct: true, message: "Nice!" };
				this._advance(this.current.idx, this.current.stage);
				this._advanceTimer = setTimeout(() => this._next(), 650);
			} else {
				this.missed[this.current.idx] = (this.missed[this.current.idx] || 0) + 1;
				this.feedback = { correct: false, message: "Not quite." };
				this._requeueSoon({ idx: this.current.idx, stage: 0 });
				this._advanceTimer = setTimeout(() => this._next(), 2200);
			}
		},

		submitWritten() {
			if (!this.current || this.current.type !== "written" || (this.feedback && !this.feedback.correct)) return;
			if (this.feedback && this.feedback.correct) {
				this._next();
				return;
			}
			this.attempts++;
			const first = this._firstTry(this.current.idx);
			this.seen[this.current.idx] = true;
			if (this._norm(this.answer) === this._norm(this.current.back)) {
				if (first) this.correctFirst++;
				this.feedback = { correct: true, message: "Correct!" };
				this._advance(this.current.idx, this.current.stage);
				this._advanceTimer = setTimeout(() => this._next(), 650);
			} else {
				this.missed[this.current.idx] = (this.missed[this.current.idx] || 0) + 1;
				this.feedback = { correct: false, message: "The correct answer is:" };
				// Missed written answers drop back to recognition.
				this._requeueSoon({ idx: this.current.idx, stage: 0 });
			}
		},

		dontKnow() {
			if (!this.current || this.feedback) return;
			this.attempts++;
			this.seen[this.current.idx] = true;
			this.missed[this.current.idx] = (this.missed[this.current.idx] || 0) + 1;
			this.feedback = { correct: false, message: "The correct answer is:" };
			if (this.current.type === "written") {
				this.answer = "";
			} else {
				this._advanceTimer = setTimeout(() => this._next(), 2200);
			}
			this._requeueSoon({ idx: this.current.idx, stage: 0 });
		},

		continue() {
			this._next();
		},

		missedList() {
			return Object.keys(this.missed).map((k) => ({
				...this._cards[Number(k)],
				count: this.missed[k],
			}));
		},

		onKey(e) {
			if (this.done || this.empty || !this.current) return;
			if (this.current.type === "mc") {
				if (!this.feedback && ["1", "2", "3", "4"].includes(e.key)) {
					const i = Number(e.key) - 1;
					if (i < this.current.options.length) this.answerMC(i);
					return;
				}
				if (e.key === "Enter" && this.feedback && !this.feedback.correct) {
					e.preventDefault();
					this.continue();
				}
				return;
			}
			if (this.current.type === "tf") {
				if (!this.feedback && (e.key === "t" || e.key === "T")) {
					this.answerTF(true);
					return;
				}
				if (!this.feedback && (e.key === "f" || e.key === "F")) {
					this.answerTF(false);
					return;
				}
				if (e.key === "Enter" && this.feedback && !this.feedback.correct) {
					e.preventDefault();
					this.continue();
				}
				return;
			}
			// written
			if (e.key === "Enter" && e.target && e.target.tagName === "INPUT") {
				e.preventDefault();
				if (this.feedback) this.continue();
				else this.submitWritten();
			} else if (e.key === "Enter" && this.feedback) {
				e.preventDefault();
				this.continue();
			}
		},
	};
}

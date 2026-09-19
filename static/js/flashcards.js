function flashcards(total) {
	return {
		currentCard: 0,
		flipped: false,
		total: total,
		prev() {
			if (this.currentCard > 0) {
				this.currentCard--;
				this.flipped = false;
			}
		},
		next() {
			if (this.currentCard < this.total - 1) {
				this.currentCard++;
				this.flipped = false;
			}
		},
		flip() {
			this.flipped = !this.flipped;
		},
		onKey(e) {
			if (e.key === "ArrowLeft") this.prev();
			else if (e.key === "ArrowRight") this.next();
			else if (e.key === "ArrowUp" || e.key === "ArrowDown" || e.key === " " || e.key === "Enter") {
				e.preventDefault();
				this.flip();
			}
		}
	};
}

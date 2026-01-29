// Card Component for Cocos Creator
// Renders individual playing cards

const { ccclass, property } = cc._decorator;

@ccclass('Card')
export default class Card extends cc.Component {
    @property(cc.Sprite)
    private cardSprite: cc.Sprite = null;

    @property(cc.Label)
    private rankLabel: cc.Label = null;

    @property(cc.Label)
    private suitLabel: cc.Label = null;

    private rank: number = 0;
    private suit: number = 0;
    private isFaceUp: boolean = false;

    onLoad() {
        this.updateCardDisplay();
    }

    setCard(rank: number, suit: number, faceUp: boolean = true) {
        this.rank = rank;
        this.suit = suit;
        this.isFaceUp = faceUp;
        this.updateCardDisplay();
    }

    setFaceUp(faceUp: boolean) {
        this.isFaceUp = faceUp;
        this.updateCardDisplay();
    }

    private updateCardDisplay() {
        if (!this.cardSprite) return;

        const ranks = ['2', '3', '4', '5', '6', '7', '8', '9', '10', 'J', 'Q', 'K', 'A'];
        const suits = ['♣', '♦', '♥', '♠'];
        const suitColors = ['black', 'red', 'red', 'black'];

        if (this.rankLabel) {
            this.rankLabel.string = this.isFaceUp ? ranks[this.rank] : '';
        }
        if (this.suitLabel) {
            this.suitLabel.string = this.isFaceUp ? suits[this.suit] : '';
            this.suitLabel.node.color = this.isFaceUp ? 
                (suitColors[this.suit] === 'red' ? cc.Color.RED : cc.Color.BLACK) : 
                cc.Color.TRANSPARENT;
        }

        // Set card back sprite when face down
        // this.cardSprite.spriteFrame = this.isFaceUp ? cardFront : cardBack;
    }

    // Card animation methods
    async dealAnimation(delay: number = 0) {
        return new Promise<void>((resolve) => {
            this.node.scale = 0;
            this.node.opacity = 0;
            
            setTimeout(() => {
                cc.tween(this.node)
                    .to(0.3, { scale: 1, opacity: 255 }, { easing: 'backOut' })
                    .call(resolve)
                    .start();
            }, delay * 1000);
        });
    }

    async flipAnimation() {
        return new Promise<void>((resolve) => {
            cc.tween(this.node)
                .to(0.15, { scaleX: 0 })
                .call(() => this.setFaceUp(!this.isFaceUp))
                .to(0.15, { scaleX: 1 })
                .call(resolve)
                .start();
        });
    }
}

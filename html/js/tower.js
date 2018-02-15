var ledstripe = {
    radius: 2,
    margin: 3,
    sep: 1,
    // columns: 0,
    // rows: 0,

    initCanvas: function () {
        this.width = this.columns * (this.radius * 2 + this.sep) - this.sep + 2 * this.margin;
        this.height = this.rows * (this.radius * 2 + this.sep) - this.sep + 2 * this.margin;
        var c = document.getElementById("ledStripeCanvas");
        c.width = this.width;
        c.height = this.height;
        var ctx = c.getContext("2d");

        ctx.beginPath();
        ctx.rect(0, 0, this.width, this.height);
        ctx.fillStyle = "#333333";
        ctx.fill();

        for (var y = 0; y < this.rows; y++) {
            for (var x = 0; x < this.columns; x++) {
                ctx.beginPath();
                ctx.arc(
                    (this.sep + 2 * this.radius) * x + this.margin + this.radius,
                    (this.sep + 2 * this.radius) * y + this.margin + this.radius,
                    this.radius, 0, 2 * Math.PI);
                ctx.strokeStyle = "white";
                ctx.stroke();
            }
        }
    },

    updateCanvas: function (data) {
        var c = document.getElementById("ledStripeCanvas");
        var ctx = c.getContext("2d");

        for (var i = 0; i < data.length; i++) {
            var y = i % this.rows;
            var x = (i - y) / this.rows;
            if (x % 2 == 1) {
                y = this.rows - 1 - y;
            }
            ctx.beginPath();
            ctx.arc(
                (this.sep + 2 * this.radius) * x + this.margin + this.radius,
                (this.sep + 2 * this.radius) * y + this.margin + this.radius,
                this.radius, 0, 2 * Math.PI);
            var rgb = data[i];
            ctx.fillStyle = '#' + rgb.toString(16).padStart(6,"0");
            ctx.fill();
        }
    }

};

function start_stripe(columns, rows) {
    ledstripe.columns = columns;
    ledstripe.rows = rows;
    ledstripe.initCanvas();
    var socket = new WebSocket("ws://127.0.0.1:8080/ws"); // TODO make it dynamic
    socket.onmessage = function (event) {
        var obj = JSON.parse(event.data);
        ledstripe.updateCanvas(obj.leds);
        // console.debug(obj);
    };
}

from flask import Flask, request, jsonify
from flask_sqlalchemy import SQLAlchemy
from flask_marshmallow import Marshmallow
from flask_cors import CORS
import os

# Inicializando o aplicativo
app = Flask(__name__)
basedir = os.path.abspath(os.path.dirname(__file__))
CORS(app)

# Configurando o banco de dados
app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///' + os.path.join(basedir, 'db.sqlite')
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

# Inicializando o banco de dados
db = SQLAlchemy(app)

# Inicializando Marshmallow
ma = Marshmallow(app)

# Modelo de Notificação
class Notification(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    type = db.Column(db.String(100))
    message = db.Column(db.String(200))

    def __init__(self, type, message):
        self.type = type
        self.message = message

# Esquema de Notificação
class NotificationSchema(ma.Schema):
    class Meta:
        fields = ('id', 'type', 'message')

# Inicializando o esquema
notification_schema = NotificationSchema()
notifications_schema = NotificationSchema(many=True)

# Criando uma Notificação
@app.route('/notification', methods=['POST'])
def add_notification():
    type = request.json['type']
    message = request.json['message']
    new_notification = Notification(type, message)
    db.session.add(new_notification)
    db.session.commit()
    return notification_schema.jsonify(new_notification)

# Obtendo todas as Notificações
@app.route('/notification', methods=['GET'])
def get_notifications():
    all_notifications = Notification.query.all()
    result = notifications_schema.dump(all_notifications)
    return jsonify(result)

# Rodando o servidor
if __name__ == '__main__':
    app.run(debug=True)

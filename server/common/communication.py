from common.utils import Bet
import logging
BET_MESSAGE_ID = "BET"
SEPARATOR = "|"
CHUNK_BET_MESSAGE_ID = 12
CHUNK_FINISH_MESSAGE_ID = 13
AGENCY_SUCCESS_MESSAGE_ID = 14
FINISH_MESSAGE_ID = 15
BYTES_MESSAGE_ID = 2
BYTES_PAYLOAD_LENGTH = 2
BYTES_CHUNK_ID_OK_MESSAGE = 4
BYTES_CLIENT_ID_FINISH_MESSAGE = 4




class AgencySocket:
    def __init__(self, socket):
        self.socket = socket

    def deserialize_chunk(self, payload: bytes) -> (str, list[Bet]):
        payload_decoded = payload.decode("utf-8")
        fields = payload_decoded.split("&")
        client_id = fields[0]
        chunk_id = fields[1]
        bets = fields[2:]
        bets_list = []
        for bet in bets:
            bet_fields = bet.split("|")
            if len(bet_fields) == 5:
                name = bet_fields[0]
                surname = bet_fields[1]
                document_id = bet_fields[2]
                birth_date = bet_fields[3]
                number = bet_fields[4]
                bet = Bet(agency=client_id, first_name=name, last_name=surname, document=document_id, birthdate=birth_date, number=number)
                bets_list.append(bet)
        return (chunk_id, bets_list)

    def receive_bets_chunk(self):
        message_id_byte = self.socket.recv(BYTES_MESSAGE_ID)
        if not message_id_byte:
            raise ConnectionError("Failed to read message ID")
        message_id = int.from_bytes(message_id_byte, byteorder='big')
        if message_id == CHUNK_BET_MESSAGE_ID:
            length_bytes = self.socket.recv(BYTES_PAYLOAD_LENGTH)
            if len(length_bytes) < BYTES_PAYLOAD_LENGTH:
                raise ConnectionError("Failed to read message length")
            msg_length = int.from_bytes(length_bytes, byteorder='big')
            payload = b""
            while len(payload) < msg_length:
                chunk = self.socket.recv(msg_length - len(payload))
                if not chunk:
                    raise ConnectionError("Connection closed before receiving full message")
                payload += chunk
            chunk_id, bets_list = self.deserialize_chunk(payload)
            return (chunk_id, bets_list)
        elif message_id == CHUNK_FINISH_MESSAGE_ID:
            client_id_bytes = self.socket.recv(BYTES_CLIENT_ID_FINISH_MESSAGE)
            if len(client_id_bytes) < BYTES_CLIENT_ID_FINISH_MESSAGE:
                raise ConnectionError("Failed to read client ID")
            client_id = int.from_bytes(client_id_bytes, byteorder='big')
            logging.info(f"FINISH REQUESTED for client id: {client_id}")

            self.send_finish_message(client_id)
            return ('', [])

        return ('', [])

    def send_ok_message(self, chunk_id):
        message_id = AGENCY_SUCCESS_MESSAGE_ID.to_bytes(BYTES_MESSAGE_ID, byteorder='big')
        chunk_id_int = int(chunk_id)
        chunk_id_bytes = chunk_id_int.to_bytes(BYTES_CHUNK_ID_OK_MESSAGE, byteorder='big')
        self.socket.sendall(message_id + chunk_id_bytes)

    def send_finish_message(self, client_id):
        message_id = FINISH_MESSAGE_ID.to_bytes(BYTES_MESSAGE_ID, byteorder='big')
        client_id = client_id.to_bytes(BYTES_CLIENT_ID_FINISH_MESSAGE, byteorder='big')
        finish_message_bytes = message_id + client_id
        self.socket.sendall(finish_message_bytes)




    def close(self):
        self.socket.close()

    def getpeername(self):
        return self.socket.getpeername()


    

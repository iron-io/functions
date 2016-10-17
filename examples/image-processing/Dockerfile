FROM iron/ruby:dev

RUN apk update && apk upgrade \
  && apk add imagemagick \
  && rm -rf /var/cache/apk/*

WORKDIR /function

ADD Gemfile* /function/
RUN bundle install

ADD . /function/

ENTRYPOINT ["ruby", "image_processor.rb"]
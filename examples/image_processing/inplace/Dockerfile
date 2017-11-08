FROM iron/ruby

RUN apk --no-cache --update add alpine-sdk gd gd-dev
RUN apk --no-cache --update add ruby ruby-io-console ruby-irb ruby-json ruby-rake ruby-rdoc ruby-dev
RUN apk --no-cache --update add imagemagick

COPY ./ /func
WORKDIR /func

RUN gem install bundler --no-ri --no-rdoc
RUN bundle

ENTRYPOINT bundle exec ruby func.rb
